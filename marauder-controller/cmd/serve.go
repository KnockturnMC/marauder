package cmd

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"time"

	"gitea.knockturnmc.com/marauder/lib/pkg/utils"

	"gitea.knockturnmc.com/marauder/controller/pkg/cronjob"
	"gitea.knockturnmc.com/marauder/controller/sqlm"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/sirupsen/logrus"

	"gitea.knockturnmc.com/marauder/controller/internal/rest"
	"github.com/gonvenience/bunt"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

func defaultConfiguration() rest.ServerConfiguration {
	return rest.ServerConfiguration{
		Host:                "localhost",
		Port:                8080,
		TLS:                 utils.TLSConfiguration{},
		KnownClientKeysFile: "{{.User.HomeDir}}/.local/marauder/controller/authorized_keys",
		Cronjobs: cronjob.CronjobsConfiguration{
			RemoveUnused: &cronjob.RemoveUnused{
				BaseCronjobConfiguration: cronjob.BaseCronjobConfiguration{
					Every: 24 * time.Hour, // run daily
				},
				RemoveAfter: 14 * 24 * time.Hour, // delete artefacts older than 14 days that are not used
			},
			RemoveHistoric: &cronjob.RemoveHistoric{
				BaseCronjobConfiguration: cronjob.BaseCronjobConfiguration{
					Every: 24 * time.Hour,
				},
				RemoveAfter: 7 * 24 * time.Hour,
			},
		},
	}
}

// ServeCommand generates the serve command for marauder controller, serving the rest server instance.
func ServeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Serves the controllers rest server",
	}

	var configurationPath string
	cmd.PersistentFlags().StringVarP(&configurationPath, "configuration", "c", "marauderctl.yml", "the path to the configuration file of the server")
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		configuration := defaultConfiguration()

		file, err := os.ReadFile(configurationPath)
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				cmd.Println(bunt.Sprint("Gray{configuration not found, using inbuilt one}"))
			} else {
				return fmt.Errorf("failed to read configuration file %s: %w", configurationPath, err)
			}
		} else {
			if err := yaml.Unmarshal(file, &configuration); err != nil {
				return fmt.Errorf("failed to parse configuration file %s: %w", configurationPath, err)
			}
		}

		dependencies, err := rest.CreateServerDependencies(version, configuration)
		if err != nil {
			return fmt.Errorf("failed to create server dependencies: %w", err)
		}

		logrus.Info("running database migrations")
		err = migrateDatabase(dependencies, configuration)
		if err != nil {
			return fmt.Errorf("failed to run migrations: %w", err)
		}

		if err := rest.StartMarauderControllerServer(configuration, dependencies); err != nil {
			return fmt.Errorf("failed to serve rest server: %w", err)
		}

		return nil
	}

	return cmd
}

func migrateDatabase(dependencies rest.ServerDependencies, configuration rest.ServerConfiguration) error {
	migrationDatabaseDriver, err := postgres.WithInstance(dependencies.DatabaseHandle.DB.DB, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("failed to create postgres migration driver for database: %w", err)
	}

	if err := sqlm.ApplyMigrations(migrationDatabaseDriver, configuration.Database.Database); err != nil {
		return fmt.Errorf("failed to apply migrations to database: %w", err)
	}

	return nil
}
