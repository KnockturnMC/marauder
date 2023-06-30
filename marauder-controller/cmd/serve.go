package cmd

import (
	"fmt"
	"os"

	"gitea.knockturnmc.com/marauder/controller/internal/rest"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

// ServeCommand generates the serve command for marauder controller, serving the rest server instance.
func ServeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Serves the controllers rest server",
	}

	var configurationPath string
	cmd.PersistentFlags().StringVarP(&configurationPath, "configuration", "c", "marauderctl.yml", "the path to the configuration file of the server")
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		file, err := os.ReadFile(configurationPath)
		if err != nil {
			return fmt.Errorf("failed to read configuration file %s: %w", configurationPath, err)
		}

		var configuration rest.ServerConfiguration
		if err := yaml.Unmarshal(file, &configuration); err != nil {
			return fmt.Errorf("failed to parse configuration file %s: %w", configurationPath, err)
		}

		if err := rest.StartMarauderControllerServer(configuration); err != nil {
			return fmt.Errorf("failed to serve rest server: %w", err)
		}

		return nil
	}

	return cmd
}
