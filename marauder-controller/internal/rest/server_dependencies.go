package rest

import (
	"bufio"
	"fmt"
	"os"

	"github.com/jmoiron/sqlx"

	"gitea.knockturnmc.com/marauder/lib/pkg/utils"

	"golang.org/x/crypto/ssh"

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

	// The ArtefactValidator used by the server to validate uploaded artefacts.
	ArtefactValidator artefact.Validator
}

// CreateServerDependencies creates the server configuration for the server based on the configuration.
func CreateServerDependencies(version string, configuration ServerConfiguration) (ServerDependencies, error) {
	artefactValidatorDispatcher, err := worker.NewDispatcher[artefact.ValidationResult](5)
	if err != nil {
		return ServerDependencies{}, fmt.Errorf("failed to create dispatcher for artefact validator: %w", err)
	}

	keys, err := parseKnownPublicKeys(configuration)
	if err != nil {
		return ServerDependencies{}, fmt.Errorf("failed to parse authorizsed keys: %w", err)
	}

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
		Version:           version,
		DatabaseHandle:    &sqlm.DB{DB: databaseHandle},
		ArtefactValidator: artefact.NewWorkedBasedValidator(artefactValidatorDispatcher, keys),
	}, nil
}

func parseKnownPublicKeys(configuration ServerConfiguration) ([]ssh.PublicKey, error) {
	authorizedKeyPath, err := utils.EvaluateFilePathTemplate(configuration.AuthorizedKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse authorized key path: %w", err)
	}

	// Parse authorized keys from disk
	authorizedKeys := make([]ssh.PublicKey, 0)
	authorizedKeysFile, err := os.Open(authorizedKeyPath)
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("failed to open authorized key file: %w", err)
		}

		return authorizedKeys, nil
	}

	defer func() { _ = authorizedKeysFile.Close() }()

	scanner := bufio.NewScanner(authorizedKeysFile)
	scanner.Split(bufio.ScanLines)

	for scanner.Scan() {
		out, _, _, _, err := ssh.ParseAuthorizedKey(scanner.Bytes())
		if err != nil {
			return nil, fmt.Errorf("failed to parse authorizsed key %s: %w", scanner.Text(), err)
		}

		authorizedKeys = append(authorizedKeys, out)
	}

	return authorizedKeys, nil
}
