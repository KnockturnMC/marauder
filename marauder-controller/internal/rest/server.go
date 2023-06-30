package rest

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"time"

	"gitea.knockturnmc.com/marauder/controller/internal/rest/v1/middleware"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// The ServerConfiguration struct holds relevant configuration values for the rest server.
type ServerConfiguration struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`

	ServerCertPath string `yaml:"serverCertPath"`
	ServerKeyPath  string `yaml:"serverKeyPath"`
}

// StartMarauderControllerServer starts the marauder controller server instance.
func StartMarauderControllerServer(configuration ServerConfiguration) error {
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
	// TODO: Register routes.

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
