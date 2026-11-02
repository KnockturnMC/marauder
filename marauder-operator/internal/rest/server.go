package rest

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/knockturnmc/marauder/marauder-lib/pkg/rest/middleware"
	"github.com/knockturnmc/marauder/marauder-operator/internal/rest/v1/endpoints"
	"github.com/sirupsen/logrus"
)

// StartMarauderOperatorServer starts the marauder operator server instance.
func StartMarauderOperatorServer(configuration ServerConfiguration, dependencies ServerDependencies) error {
	ginRootServer := gin.New()
	if err := ginRootServer.SetTrustedProxies(nil); err != nil {
		return fmt.Errorf("failed to set server trusted proxies: %w", err)
	}

	logrus.Debug("registering middleware on gin server")
	ginRootServer.Use(gin.LoggerWithFormatter(middleware.RequestLoggerFormatter()))
	ginRootServer.Use(gin.Recovery())
	ginRootServer.Use(cors.Default())
	ginRootServer.Use(middleware.ErrorHandler())

	logrus.Debug("registering routs on gin server")
	group := ginRootServer.Group("/v1")
	group.GET("/version", endpoints.VersionGet(dependencies.Version))
	group.POST("/cron/cache/clear", endpoints.CronCleanCache(dependencies.DownloadingService))

	serverWithUUID := group.Group("/server/:uuid").Use(endpoints.ServerUUIDMiddleware(
		configuration.Identifier,
		dependencies.ControllerClient,
	))
	serverWithUUID.POST("/:action", endpoints.ServerLifecycleActionPost(
		dependencies.ServerManager,
	))

	serverWithUUID.GET("/management/players", endpoints.ServerManagementPlayers(
		dependencies.ServerManager,
	))
	serverWithUUID.POST("/management/togglesave", endpoints.ServerManagementToggleSave(
		dependencies.ServerManager,
	))
	serverWithUUID.GET("/system/image/details", endpoints.ServerSystemImageDetails(
		dependencies.ServerManager,
	))

	logrus.Info("staring server on port ", configuration.Port)
	engine := &http.Server{
		Addr:              fmt.Sprintf("%s:%d", configuration.Host, configuration.Port),
		Handler:           ginRootServer,
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
