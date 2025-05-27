package rest

import (
	"crypto/tls"
	"fmt"
	"net/http"

	"github.com/jmoiron/sqlx"
	"github.com/knockturnmc/marauder/marauder-controller/internal/cronjobworker"
	"github.com/knockturnmc/marauder/marauder-controller/pkg/artefact"
	"github.com/knockturnmc/marauder/marauder-controller/sqlm"
	"github.com/knockturnmc/marauder/marauder-lib/pkg/keyauth"
	"github.com/knockturnmc/marauder/marauder-lib/pkg/operator"
	"github.com/knockturnmc/marauder/marauder-lib/pkg/utils"
	"github.com/knockturnmc/marauder/marauder-lib/pkg/worker"
	_ "github.com/lib/pq" // postgres driver
	"github.com/sirupsen/logrus"
)

// The ServerDependencies holds all state and instances needed for the rest server to function.
type ServerDependencies struct {
	// The version of the server.
	Version string

	// The DatabaseHandle to marauderctl's database.
	DatabaseHandle *sqlm.DB

	// The ArtefactValidator used by the server to validate uploaded artefacts.
	ArtefactValidator artefact.Validator

	// CronjobWorker is the cronjob worker the controller server uses.
	CronjobWorker *cronjobworker.CronjobWorker

	// OperatorClientCache holds the http clients to communicate with the operators managed by the controller.
	OperatorClientCache *operator.ClientCache

	// The TLSConfig for the server if tls is enabled.
	TLSConfig *tls.Config
}

// CreateServerDependencies creates the server configuration for the server based on the configuration.
func CreateServerDependencies(version string, configuration ServerConfiguration) (ServerDependencies, error) {
	artefactValidatorDispatcher, err := worker.NewDispatcher[artefact.ValidationResult](5)
	if err != nil {
		return ServerDependencies{}, fmt.Errorf("failed to create dispatcher for artefact validator: %w", err)
	}

	logrus.Debug("looking for local tls configuration")
	tlsConfiguration, err := utils.ParseTLSConfigurationFromType(configuration.TLS)
	if err != nil {
		logrus.Warnf("failed to enable tsl: %s", err)
	}

	logrus.Debug("loading known public keys of artefact signers")
	keys, err := keyauth.ParseKnownPublicKeys(configuration.KnownClientKeysFile)
	if err != nil {
		return ServerDependencies{}, fmt.Errorf("failed to parse authorizsed keys: %w", err)
	}

	logrus.Debug("connecting to database")
	databaseConnectionString := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable binary_parameters=yes",
		configuration.Database.Host, configuration.Database.Port, configuration.Database.Username,
		configuration.Database.Password, configuration.Database.Database,
	)
	databaseHandle, err := sqlx.Connect("postgres", databaseConnectionString)
	if err != nil {
		return ServerDependencies{}, fmt.Errorf("failed to open database connection pool: %w", err)
	}

	wrappedDatabaseHandle := &sqlm.DB{DB: databaseHandle}

	operatorClientCacheSharedHTTPClient := &http.Client{}
	protocol := "http"
	if tlsConfiguration != nil {
		protocol = "https"
	}
	operatorClientCache := operator.NewOperatorClientCache(operatorClientCacheSharedHTTPClient, protocol)

	// tls is enabled
	if tlsConfiguration != nil {
		operatorClientCacheSharedHTTPClient.Transport = &http.Transport{ // Enable on client for operator
			TLSClientConfig: tlsConfiguration.Clone(),
		}

		// Configure client auth requirement for server side tls config.
		tlsConfiguration.ClientAuth = tls.RequireAndVerifyClientCert
		tlsConfiguration.ClientCAs = tlsConfiguration.RootCAs
	}

	cronjobWorker := cronjobworker.NewCronjobWorker(
		wrappedDatabaseHandle,
		operatorClientCache,
		cronjobworker.ComputeCronjobMap(configuration.Cronjobs),
	)

	return ServerDependencies{
		Version:             version,
		DatabaseHandle:      wrappedDatabaseHandle,
		ArtefactValidator:   artefact.NewWorkedBasedValidator(artefactValidatorDispatcher, keys),
		OperatorClientCache: operatorClientCache,
		CronjobWorker:       cronjobWorker,
		TLSConfig:           tlsConfiguration,
	}, nil
}
