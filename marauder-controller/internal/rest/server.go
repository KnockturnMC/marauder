package rest

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"gitea.knockturnmc.com/marauder/lib/pkg/utils"

	"github.com/ztrue/shutdown"

	"gitea.knockturnmc.com/marauder/controller/pkg/cronjob"

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

	Cronjobs cronjob.CronjobsConfiguration `yaml:"cronjobs"`

	TLS utils.TLSConfiguration `yaml:"tls"`

	KnownClientKeysFile string `yaml:"knownClientKeysFile"`
}

// StartMarauderControllerServer starts the marauder controller server instance.
func StartMarauderControllerServer(configuration ServerConfiguration, dependencies ServerDependencies) error {
	server := gin.New()
	if err := server.SetTrustedProxies(nil); err != nil {
		return fmt.Errorf("failed to set server trusted proxies: %w", err)
	}

	configureRouterGroup(server, dependencies)

	logrus.Info("staring server on port ", configuration.Port)
	engine := &http.Server{
		Addr:              fmt.Sprintf("%s:%d", configuration.Host, configuration.Port),
		Handler:           server,
		ReadHeaderTimeout: 30 * time.Second,
		TLSConfig:         dependencies.TLSConfig,
	}

	startCronjobWorker(dependencies)

	var serveErr error
	if engine.TLSConfig != nil {
		serveErr = engine.ListenAndServeTLS("", "") // Defined in config
	} else {
		serveErr = engine.ListenAndServe()
	}
	if serveErr != nil {
		return fmt.Errorf("failed to listen and serve: %w", serveErr)
	}

	return nil
}

// startCronjobWorker starts the cronjob worker passed in the server dependencies.
func startCronjobWorker(dependencies ServerDependencies) {
	cronjobWorkerContext, cronjobWorkerCancel := context.WithCancel(context.Background())
	go func() {
		if err := dependencies.CronjobWorker.Start(cronjobWorkerContext); err != nil {
			logrus.Errorf("failed cronjob worker: %s", err)
			return
		}
	}()
	shutdown.Add(cronjobWorkerCancel) // shutdown worker on shutdown
}

// configureRouterGroup configures the router for the engine, specifically all its endpoints.
func configureRouterGroup(server *gin.Engine, dependencies ServerDependencies) {
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
	group.GET("/server/:uuid/state/update", endpoints.ServerUpdate(dependencies.DatabaseHandle))
	group.GET("/server/:uuid/state/:state", endpoints.ServerStateGet(dependencies.DatabaseHandle))
	group.PATCH("/server/:uuid/state/:state", endpoints.ServerDeploymentPatch(dependencies.DatabaseHandle))
	group.DELETE("/server/:uuid/state/:state", endpoints.ServerDeploymentPatch(dependencies.DatabaseHandle))

	operatorProtocol := "http"
	if dependencies.TLSConfig != nil {
		operatorProtocol = "https"
	}
	group.Any("/operator/:server/*path", endpoints.OperationServerProxy(
		dependencies.DatabaseHandle,
		dependencies.OperatorHTTPClient,
		operatorProtocol,
	))
}
