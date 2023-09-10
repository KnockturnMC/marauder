package cmd

import (
	"errors"
	"fmt"
	"net/http"
	"os"

	"gitea.knockturnmc.com/marauder/lib/pkg/controller"

	"gitea.knockturnmc.com/marauder/lib/pkg/utils"
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
func (c Configuration) CreateTLSReadyHTTPClient() (controller.Client, error) {
	configuration, err := utils.ParseTLSConfigurationFromType(c.TLS)
	if err != nil {
		return &controller.HTTPClient{
			Client:        http.DefaultClient,
			ControllerURL: c.ControllerHost,
		}, fmt.Errorf("failed to parse tls config: %w", err)
	}

	return &controller.HTTPClient{
		Client: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: configuration,
			},
		},
		ControllerURL: c.ControllerHost,
	}, nil
}

// ParseSigningKey parses the signing key as defined int the configuration.
func (c Configuration) ParseSigningKey() (ssh.Signer, error) {
	privateKeyFilePath, err := utils.EvaluateFilePathTemplate(c.SigningKey)
	if err != nil {
		return nil, fmt.Errorf("failed to evaluate private key file path: %w", err)
	}

	privateKeyBytes, err := os.ReadFile(privateKeyFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read private key file for signing: %w", err)
	}

	key, err := ssh.ParsePrivateKey(privateKeyBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key bytes: %w", err)
	}

	return key, nil
}
