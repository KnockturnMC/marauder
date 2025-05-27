package rest

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"os"

	dockerClient "github.com/docker/docker/client"
	"github.com/knockturnmc/marauder/marauder-lib/pkg/controller"
	"github.com/knockturnmc/marauder/marauder-lib/pkg/fileeq"
	"github.com/knockturnmc/marauder/marauder-lib/pkg/utils"
	"github.com/knockturnmc/marauder/marauder-lib/pkg/worker"
	"github.com/knockturnmc/marauder/marauder-operator/pkg/manager"
	_ "github.com/lib/pq" // postgres driver
	"github.com/sirupsen/logrus"
)

// The ServerDependencies holds all state and instances needed for the rest server to function.
type ServerDependencies struct {
	// The version of the server.
	Version string

	// The http client to communicate with the operator
	ControllerClient controller.DownloadingClient

	// The used downloading service, cleanable by request.
	DownloadingService worker.DownloadService

	// The ServerManager is responsible for managing the docker instances on the server.
	ServerManager manager.Manager

	// TLSConfig provides the tsl configuration for the gin engine.
	TLSConfig *tls.Config
}

// CreateServerDependencies creates the server configuration for the server based on the configuration.
func CreateServerDependencies(version string, configuration ServerConfiguration) (ServerDependencies, error) {
	logrus.Debug("looking for local tls configuration")
	tlsConfiguration, err := utils.ParseTLSConfigurationFromType(configuration.TLS)
	if err != nil {
		logrus.Warnf("failed to enable tsl: %s", err)
	}

	logrus.Debug("creating downloads folder on disk")
	if err := os.MkdirAll(configuration.Disk.DownloadPath, 0o700); err != nil {
		return ServerDependencies{}, fmt.Errorf("failed to create download path for marauder operator: %w", err)
	}

	logrus.Debug("creating docker client")
	dockerClientInstance, err := dockerClient.NewClientWithOpts(dockerClient.FromEnv, dockerClient.WithAPIVersionNegotiation())
	if err != nil {
		return ServerDependencies{}, fmt.Errorf("failed to create docker client: %w", err)
	}

	dockerEncodedBasicAuth, err := configuration.Docker.ToBasicAuth()
	if err != nil {
		return ServerDependencies{}, fmt.Errorf("failed to encode docker authentication: %w", err)
	}

	controllerHTTPClient := &http.Client{}

	// tls is enabled
	if tlsConfiguration != nil {
		controllerHTTPClient.Transport = &http.Transport{
			TLSClientConfig: tlsConfiguration.Clone(),
		}

		tlsConfiguration.ClientAuth = tls.RequireAndVerifyClientCert
		tlsConfiguration.ClientCAs = tlsConfiguration.RootCAs
	}

	dispatcher, err := worker.NewDispatcher[worker.DownloadResult](configuration.Controller.WorkerCount)
	if err != nil {
		return ServerDependencies{}, fmt.Errorf("failed to create dispatcher for controller client: %w", err)
	}

	downloadService := worker.NewMutexDownloadService(controllerHTTPClient, dispatcher, configuration.Disk.DownloadPath)

	controllerClient := &controller.DownloadingHTTPClient{
		HTTPClient: controller.HTTPClient{
			Client:        controllerHTTPClient,
			ControllerURL: configuration.Controller.Endpoint,
		},
		DownloadService: downloadService,
	}

	return ServerDependencies{
		Version:            version,
		ControllerClient:   controllerClient,
		TLSConfig:          tlsConfiguration,
		DownloadingService: downloadService,
		ServerManager: &manager.DockerBasedManager{
			ControllerClient:       controllerClient,
			DockerClient:           dockerClientInstance,
			DockerEncodedAuth:      dockerEncodedBasicAuth,
			AutoRemoveContainers:   configuration.Docker.AutoRemoveContainers,
			ContainerMemoryBuffer:  configuration.Docker.ContainerMemoryBuffer,
			FolderOwner:            configuration.Disk.FolderOwner,
			ServerDataPathTemplate: configuration.Disk.ServerDataPathTemplate,
			FileEqualityRegistry:   fileeq.DefaultFileEqualityRegistry(),
		},
	}, nil
}
