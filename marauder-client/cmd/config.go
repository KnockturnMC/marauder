package cmd

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/knockturnmc/marauder/marauder-lib/pkg/controller"
	"github.com/knockturnmc/marauder/marauder-lib/pkg/utils"
	"github.com/knockturnmc/marauder/marauder-lib/pkg/worker"
	"golang.org/x/crypto/ssh"
)

type CommandContextKeyType int

const (
	// KeyBuildCommandTarballOutput defines the shared key used by the build command to store the build output target in a context.
	KeyBuildCommandTarballOutput CommandContextKeyType = iota
	KeyPublishResultArtefactModel
)

// ErrContextMissingValue is returned if a command expects a context to contain a specific value, but it isn't there.
var ErrContextMissingValue = errors.New("context missing value")

// DefaultConfiguration defines the default configuration.
func DefaultConfiguration() Configuration {
	defaultTLSFolder := "{{.User.HomeDir}}/.local/marauder/client/tls"
	return Configuration{
		ControllerHost: "http://localhost:8080/v1",
		TLS: utils.TLSConfiguration{
			Folder: &defaultTLSFolder,
		},
		SigningKey: "{{.User.HomeDir}}/.local/marauder/client/signingKey",
	}
}

// The Configuration type represents the configuration of the client cli.
type Configuration struct {
	ControllerHost string                 `yaml:"controllerHost"`
	TLS            utils.TLSConfiguration `yaml:"tls"`
	SigningKey     string                 `yaml:"signingKey"`
}

// CreateTLSReadyHTTPClient creates a tls ready http client for communication with the controller.
func (c Configuration) CreateTLSReadyHTTPClient() (controller.DownloadingClient, error) {
	// Create download dir
	cacheDir, err := os.MkdirTemp("", "marauder-client-cache")
	if err != nil {
		return nil, fmt.Errorf("failed to create download cache: %w", err)
	}

	dispatcher, err := worker.NewDispatcher[worker.DownloadResult](1)
	if err != nil {
		return nil, fmt.Errorf("failed to create dispatcher for download client: %w", err)
	}

	configuration, err := utils.ParseTLSConfigurationFromType(c.TLS)
	if err != nil {
		return &controller.DownloadingHTTPClient{
			HTTPClient: controller.HTTPClient{
				Client:        http.DefaultClient,
				ControllerURL: c.ControllerHost,
			},
			DownloadService: worker.NewMutexDownloadService(http.DefaultClient, dispatcher, cacheDir),
		}, fmt.Errorf("failed to parse tls config: %w", err)
	}

	httpClient := &http.Client{Transport: &http.Transport{TLSClientConfig: configuration}}
	tlsDownloadService := worker.NewMutexDownloadService(httpClient, dispatcher, cacheDir)
	return &controller.DownloadingHTTPClient{
		HTTPClient: controller.HTTPClient{
			Client:        httpClient,
			ControllerURL: c.ControllerHost,
		},
		DownloadService: tlsDownloadService,
	}, nil
}

// ParseSigningKey parses the signing key as defined int the configuration.
func (c Configuration) ParseSigningKey() (ssh.Signer, error) {
	privateKeyFilePath, err := utils.EvaluateFilePathTemplate(c.SigningKey)
	if err != nil {
		return nil, fmt.Errorf("failed to evaluate private key file path: %w", err)
	}

	privateKeyBytes, err := os.ReadFile(filepath.Clean(privateKeyFilePath))
	if err != nil {
		return nil, fmt.Errorf("failed to read private key file for signing: %w", err)
	}

	key, err := ssh.ParsePrivateKey(privateKeyBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key bytes: %w", err)
	}

	return key, nil
}
