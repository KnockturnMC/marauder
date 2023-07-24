package rest

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"time"

	"gitea.knockturnmc.com/marauder/operator/internal/rest/v1/endpoints"

	"gitea.knockturnmc.com/marauder/lib/pkg/rest/middleware"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// The ServerConfiguration struct holds relevant configuration values for the rest server.
type ServerConfiguration struct {
	Identifier string `yaml:"identifier"`

	Host string `yaml:"host"`
	Port int    `yaml:"port"`

	ControllerEndpoint string `yaml:"controllerEndpoint"`

	Disk Disk `yaml:"disk"`

	ServerCertPath string `yaml:"serverCertPath"`
	ServerKeyPath  string `yaml:"serverKeyPath"`
}

// Disk contains configuration values for the disk setup of controller.
type Disk struct {
	DownloadPath           string `yaml:"downloadPath"`
	ServerDataPathTemplate string `yaml:"serverDataPathTemplate"`
}

// StartMarauderOperatorServer starts the marauder operator server instance.
func StartMarauderOperatorServer(configuration ServerConfiguration, dependencies ServerDependencies) error {
	server := gin.New()
	if err := server.SetTrustedProxies(nil); err != nil {
		return fmt.Errorf("failed to set server trusted proxies: %w", err)
	}

	logrus.Debug("registering middleware on gin server")
	server.Use(gin.LoggerWithFormatter(middleware.RequestLoggerFormatter()))
	server.Use(gin.Recovery())
	server.Use(cors.Default())
	server.Use(middleware.ErrorHandler())

	logrus.Debug("registering routs on gin server")
	group := server.Group("/v1")
	group.GET("/version", endpoints.VersionGet(dependencies.Version))

	group.POST("/server/:uuid/:action", endpoints.ServerLifecycleActionPost(
		configuration.Identifier,
		dependencies.ControllerClient,
		dependencies.ServerManager,
	))

	logrus.Info("staring server on port ", configuration.Port)
	engine := &http.Server{
		Addr:              fmt.Sprintf("%s:%d", configuration.Host, configuration.Port),
		Handler:           server,
		ReadHeaderTimeout: 30 * time.Second,
		TLSConfig: &tls.Config{
			MinVersion: tls.VersionTLS13,
		},
	}

	serverListenAndServeErr := engine.ListenAndServe()
	if serverListenAndServeErr != nil {
		return fmt.Errorf("failed to listen and serve: %w", serverListenAndServeErr)
	}

	return nil
}
