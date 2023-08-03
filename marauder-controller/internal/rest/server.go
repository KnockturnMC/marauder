package rest

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"time"

	middleware2 "gitea.knockturnmc.com/marauder/lib/pkg/rest/middleware"

	"gitea.knockturnmc.com/marauder/controller/internal/rest/v1/endpoints"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// The ServerConfiguration struct holds relevant configuration values for the rest server.
type ServerConfiguration struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`

	Database struct {
		Host     string `yaml:"host"`
		Port     int    `yaml:"port"`
		Username string `yaml:"username"`
		Password string `yaml:"password"`
		Database string `yaml:"database"`
	} `yaml:"database"`

	ServerCertPath    string `yaml:"serverCertPath"`
	ServerKeyPath     string `yaml:"serverKeyPath"`
	AuthorizedKeyPath string `yaml:"authorizedKeyPath"`
}

// StartMarauderControllerServer starts the marauder controller server instance.
func StartMarauderControllerServer(configuration ServerConfiguration, dependencies ServerDependencies) error {
	server := gin.New()
	if err := server.SetTrustedProxies(nil); err != nil {
		return fmt.Errorf("failed to set server trusted proxies: %w", err)
	}

	logrus.Debug("registering middleware on gin server")
	server.Use(gin.LoggerWithFormatter(middleware2.RequestLoggerFormatter()))
	server.Use(gin.Recovery())
	server.Use(cors.Default())
	server.Use(middleware2.ErrorHandler())

	logrus.Debug("registering routs on gin server")
	group := server.Group("/v1")
	group.GET("/version", endpoints.VersionGet(dependencies.Version))

	group.POST("/artefact", endpoints.ArtefactUploadGet(dependencies.DatabaseHandle, dependencies.ArtefactValidator))
	group.GET("/artefact/:uuid", endpoints.ArtefactUUIDGet(dependencies.DatabaseHandle))
	group.GET("/artefact/:uuid/download", endpoints.ArtefactUUIDDownloadGet(dependencies.DatabaseHandle))
	group.GET("/artefact/:uuid/download/manifest", endpoints.ArtefactUUIDDownloadManifestGet(dependencies.DatabaseHandle))
	group.GET("/artefacts/:identifier", endpoints.ArtefactsIdentifierGet(dependencies.DatabaseHandle))
	group.GET("/artefacts/:identifier/:version", endpoints.ArtefactIdentifierVersionGet(dependencies.DatabaseHandle))

	group.GET("/server/:uuid", endpoints.ServerUUIDGet(dependencies.DatabaseHandle))
	group.GET("/servers/:environment", endpoints.ServersEnvironmentGet(dependencies.DatabaseHandle))
	group.GET("/servers/:environment/:name", endpoints.ServersEnvironmentNameGet(dependencies.DatabaseHandle))

	group.GET("/server/:uuid/state/", endpoints.ServerStateGet(dependencies.DatabaseHandle))
	group.GET("/server/:uuid/state/:state", endpoints.ServerStateGet(dependencies.DatabaseHandle))
	group.PATCH("/server/:uuid/state/:state", endpoints.ServerDeploymentPatch(dependencies.DatabaseHandle))
	group.GET("/server/:uuid/apply/update", endpoints.ServerUpdate(dependencies.DatabaseHandle))

	group.Any("/operator/:server/*path", endpoints.OperationServerProxy(dependencies.DatabaseHandle, dependencies.OperatorHTTPClient))

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
