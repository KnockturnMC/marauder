package cmd

import (
	"fmt"
	"net/http"
	"os"

	"gitea.knockturnmc.com/marauder/lib/pkg/utils"
	"golang.org/x/crypto/ssh"
)

type CommandContextKeyType int

const (
	// KeyBuildCmdOutput defines the shared key used by the build command to store the build output target in a context.
	KeyBuildCmdOutput CommandContextKeyType = iota
)

// DefaultConfiguration defines the default configuration.
func DefaultConfiguration() Configuration {
	return Configuration{
		ControllerHost: "http://localhost:8080/v1",
		TLSPath:        "{{.User.HomeDir}}/.config/marauder/client/tls",
		SigningKey:     "{{.User.HomeDir}}/.config/marauder/client/signingKey",
	}
}

// The Configuration type represents the configuration of the client cli.
type Configuration struct {
	ControllerHost string `yaml:"controllerHost"`
	TLSPath        string `yaml:"tlsPath"`
	SigningKey     string `yaml:"signingKey"`
}

// CreateTLSReadyHTTPClient creates a tls ready http client for communication with the controller.
func (c Configuration) CreateTLSReadyHTTPClient() (*http.Client, error) {
	configuration, err := utils.ParseTLSConfiguration(c.TLSPath)
	if err != nil {
		return http.DefaultClient, fmt.Errorf("failed to parse tls config: %w", err)
	}

	return &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: configuration,
		},
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
