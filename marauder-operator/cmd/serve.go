package cmd

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/gonvenience/bunt"
	"github.com/knockturnmc/marauder/marauder-lib/pkg/utils"
	"github.com/knockturnmc/marauder/marauder-operator/internal/rest"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

func defaultConfiguration() rest.ServerConfiguration {
	return rest.ServerConfiguration{
		Identifier: "localhost",
		Host:       "localhost",
		Port:       1981,
		TLS:        utils.TLSConfiguration{},
		Docker: rest.Docker{
			Username:              "",
			Password:              "",
			AutoRemoveContainers:  true,
			ContainerMemoryBuffer: 512, // in MB
		},
		Controller: rest.Controller{
			Endpoint:    "http://localhost:8080/v1",
			WorkerCount: 5,
		},
		Disk: rest.Disk{
			DownloadPath:           "/var/local/marauder/operator/cache/downloads",
			ServerDataPathTemplate: "/var/local/marauder/operator/servers/{{.Environment}}/{{.Name}}",
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
	cmd.PersistentFlags().StringVarP(&configurationPath, "configuration", "c", "marauderop.yml", "the path to the configuration file of the server")
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		configuration := defaultConfiguration()

		file, err := os.ReadFile(filepath.Clean(configurationPath))
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

		if err := rest.StartMarauderOperatorServer(configuration, dependencies); err != nil {
			return fmt.Errorf("failed to serve rest server: %w", err)
		}

		return nil
	}

	return cmd
}
