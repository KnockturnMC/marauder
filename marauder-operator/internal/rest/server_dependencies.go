package rest

import (
	"fmt"
	"net/http"
	"os"

	_ "github.com/lib/pq" // postgres driver
	"github.com/sirupsen/logrus"
)

// The ServerDependencies holds all state and instances needed for the rest server to function.
type ServerDependencies struct {
	// The version of the server.
	Version string

	// The http client to communicate with the operator
	ControllerClient http.Client
}

// CreateServerDependencies creates the server configuration for the server based on the configuration.
func CreateServerDependencies(version string, configuration ServerConfiguration) (ServerDependencies, error) {
	logrus.Debug("creating downloads folder on disk")
	if err := os.MkdirAll(configuration.Disk.DownloadPath, 0o700); err != nil {
		return ServerDependencies{}, fmt.Errorf("failed to create download path for marauder operator: %w", err)
	}

	logrus.Debug("creating server data folder on disk")
	if err := os.MkdirAll(configuration.Disk.ServerDataPath, 0o700); err != nil {
		return ServerDependencies{}, fmt.Errorf("failed to create server data path for marauder operator: %w", err)
	}

	return ServerDependencies{
		Version:          version,
		ControllerClient: http.Client{},
	}, nil
}
