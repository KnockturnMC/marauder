package cmd

import (
	"fmt"
	"os"

	"gitea.knockturnmc.com/marauder/lib/pkg/utils"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var version = "develop"

// RootCommand is the root entry command for the builder tool.
func RootCommand(configuration *Configuration) *cobra.Command {
	var configurationPath string

	command := &cobra.Command{
		Use:   "marauder",
		Short: "Marauder is a command line tool capable of constructing artefacts",
		Long: `Marauder is a command line tool capable of packing together a locally defined artefact into a
tarball and uploading said artefact to the marauder controller.`,
		Version: version,
	}
	command.PersistentFlags().StringVarP(
		&configurationPath,
		"configPath",
		"c",
		"{{.User.HomeDir}}/.config/marauder/client/config.yml",
		"the path on the host to the marauder client configuration",
	)

	command.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		resolvedConfigurationPath, err := utils.EvaluateFilePathTemplate(configurationPath)
		if err != nil {
			return fmt.Errorf("failed to execute configuration path template: %w", err)
		}

		configurationBytes, err := os.ReadFile(resolvedConfigurationPath)
		if err != nil {
			return nil //nolint:nilerr
		}

		if err := yaml.Unmarshal(configurationBytes, configuration); err != nil {
			return fmt.Errorf("failed to unmarshal configuration content: %w", err)
		}

		return nil
	}

	return command
}
