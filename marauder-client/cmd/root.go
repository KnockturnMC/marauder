package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/gonvenience/bunt"

	"gitea.knockturnmc.com/marauder/lib/pkg/utils"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// ErrIncorrectArgumentFormat is returned if the argument is in a wrong format.
var ErrIncorrectArgumentFormat = errors.New("incorrect format")

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
		configurationBytes, err := ReadFileFromOrStdin(configurationPath, cmd.InOrStdin())
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

// ReadFileFromOrStdin reads the contents found at the path and returns them or, if the path is "-", reads the entire stdin and yields
// them back.
func ReadFileFromOrStdin(path string, stdin io.Reader) ([]byte, error) {
	if path == "-" {
		stdinBytes, err := io.ReadAll(stdin)
		if err != nil {
			return nil, fmt.Errorf("failed to read stdin from pathOrStdin: %w", err)
		}

		return stdinBytes, nil
	}

	resolvedPath, err := utils.EvaluateFilePathTemplate(path)
	if err != nil {
		return nil, fmt.Errorf("failed to execute path template: %w", err)
	}

	fileBytes, err := os.ReadFile(resolvedPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", path, err)
	}

	return fileBytes, nil
}

// printArtefactFetchResult prints the passed result set to the command output stream.
func printFetchResult[R any](cmd *cobra.Command, result R) {
	output, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		cmd.PrintErrln(bunt.Sprintf("Red{failed to marshal result %s}", err))
		return
	}

	cmd.Println(string(output))
}
