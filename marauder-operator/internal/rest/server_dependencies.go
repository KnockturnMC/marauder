package rest

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"os"

	"gitea.knockturnmc.com/marauder/lib/pkg/worker"

	"gitea.knockturnmc.com/marauder/operator/pkg/servermgr"
	dockerClient "github.com/docker/docker/client"

	"gitea.knockturnmc.com/marauder/operator/pkg/controller"

	_ "github.com/lib/pq" // postgres driver
	"github.com/sirupsen/logrus"
)

// The ServerDependencies holds all state and instances needed for the rest server to function.
type ServerDependencies struct {
	// The version of the server.
	Version string

	// The http client to communicate with the operator
	ControllerClient controller.Client

	// The ServerManager is responsible for managing the docker instances on the server.
	ServerManager servermgr.Manager

	// TLSConfig provides the tsl configuration for the gin engine.
	TLSConfig *tls.Config
}

// CreateServerDependencies creates the server configuration for the server based on the configuration.
func CreateServerDependencies(version string, configuration ServerConfiguration) (ServerDependencies, error) {
	logrus.Debug("creating downloads folder on disk")
	if err := os.MkdirAll(configuration.Disk.DownloadPath, 0o700); err != nil {
		return ServerDependencies{}, fmt.Errorf("failed to create download path for marauder operator: %w", err)
	}

	logrus.Debug("creating docker client")
	dockerClientInstance, err := dockerClient.NewClientWithOpts(dockerClient.FromEnv, dockerClient.WithAPIVersionNegotiation())
	if err != nil {
		return ServerDependencies{}, fmt.Errorf("failed to create docker client: %w", err)
	}

	httpClient := &http.Client{}

	dispatcher, err := worker.NewDispatcher[worker.DownloadResult](configuration.Controller.WorkerCount)
	if err != nil {
		return ServerDependencies{}, fmt.Errorf("failed to create dispatcher for controller client: %w", err)
	}

	downloadService := worker.NewMutexDownloadService(httpClient, dispatcher, configuration.Disk.DownloadPath)

	return ServerDependencies{
		Version: version,
		ControllerClient: &controller.HTTPClient{
			Client:          httpClient,
			ControllerURL:   configuration.Controller.Endpoint,
			DownloadService: downloadService,
		},
		ServerManager: &servermgr.DockerBasedManager{
			DockerClient:           dockerClientInstance,
			ServerDataPathTemplate: configuration.Disk.ServerDataPathTemplate,
		},
	}, nil
}
