package rest

import (
	"crypto/tls"
	"fmt"
	"net/http"

	"gitea.knockturnmc.com/marauder/controller/internal/cronjobworker"

	"gitea.knockturnmc.com/marauder/lib/pkg/utils"

	"gitea.knockturnmc.com/marauder/lib/pkg/keyauth"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"

	"gitea.knockturnmc.com/marauder/controller/pkg/artefact"
	"gitea.knockturnmc.com/marauder/controller/sqlm"
	"gitea.knockturnmc.com/marauder/lib/pkg/worker"
	_ "github.com/lib/pq" // postgres driver
)

// The ServerDependencies holds all state and instances needed for the rest server to function.
type ServerDependencies struct {
	// The version of the server.
	Version string

	// The DatabaseHandle to marauderctl's database.
	DatabaseHandle *sqlm.DB

	// The OperatorHTTPClient is a http client with a tls configuration authorized to communicate with
	// operator instances.
	OperatorHTTPClient *http.Client

	// The ArtefactValidator used by the server to validate uploaded artefacts.
	ArtefactValidator artefact.Validator

	// JobWorker is the job worker the controller server uses.
	JobWorker cronjobworker.CronjobWorker

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
	tlsConfiguration, err := utils.ParseTLSConfiguration(configuration.TLSPath)
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

	operatorClient := &http.Client{}

	// tls is enabled
	if tlsConfiguration != nil {
		operatorClient.Transport = &http.Transport{ // Enable on client for operator
			TLSClientConfig: tlsConfiguration.Clone(),
		}

		// Configure client auth requirement for server side tls config.
		tlsConfiguration.ClientAuth = tls.RequireAndVerifyClientCert
		tlsConfiguration.ClientCAs = tlsConfiguration.RootCAs
	}

	return ServerDependencies{
		Version:            version,
		DatabaseHandle:     &sqlm.DB{DB: databaseHandle},
		OperatorHTTPClient: operatorClient,
		ArtefactValidator:  artefact.NewWorkedBasedValidator(artefactValidatorDispatcher, keys),
		TLSConfig:          tlsConfiguration,
	}, nil
}
