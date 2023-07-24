package rest

import (
	"fmt"
	"net/http"

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
}

// CreateServerDependencies creates the server configuration for the server based on the configuration.
func CreateServerDependencies(version string, configuration ServerConfiguration) (ServerDependencies, error) {
	artefactValidatorDispatcher, err := worker.NewDispatcher[artefact.ValidationResult](5)
	if err != nil {
		return ServerDependencies{}, fmt.Errorf("failed to create dispatcher for artefact validator: %w", err)
	}

	logrus.Debug("loading known public keys of artefact signers")
	keys, err := keyauth.ParseKnownPublicKeys(configuration.AuthorizedKeyPath)
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

	return ServerDependencies{
		Version:            version,
		DatabaseHandle:     &sqlm.DB{DB: databaseHandle},
		OperatorHTTPClient: &http.Client{},
		ArtefactValidator:  artefact.NewWorkedBasedValidator(artefactValidatorDispatcher, keys),
	}, nil
}
