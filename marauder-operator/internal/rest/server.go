package rest

import (
	"fmt"
	"net/http"
	"time"

	"gitea.knockturnmc.com/marauder/operator/internal/rest/v1/endpoints"

	"gitea.knockturnmc.com/marauder/lib/pkg/rest/middleware"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

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
	group.POST("/cron/cache/clear", endpoints.CronCleanCache(dependencies.DownloadingService))

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
		TLSConfig:         dependencies.TLSConfig,
	}

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
