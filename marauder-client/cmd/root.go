package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"

	"gitea.knockturnmc.com/marauder/lib/pkg/controller"

	"gitea.knockturnmc.com/marauder/lib/pkg/models/networkmodel"
	"github.com/google/uuid"

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

// printArtefactFetchResult prints the passed result set to the command output stream.
func printFetchResult[R any](cmd *cobra.Command, result R) {
	output, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		cmd.PrintErrln(bunt.Sprintf("Red{failed to marshal result %s}", err))
		return
	}

	cmd.Println(string(output))
}

// fetchArtefact fetches a specific artefact from the remote controller host.
func fetchArtefact(
	ctx context.Context,
	client controller.Client,
	configuration *Configuration,
	artefactUUID uuid.UUID,
) (networkmodel.ArtefactModel, error) {
	if err != nil {
		return networkmodel.ArtefactModel{}, fmt.Errorf("failed to fetch artefact: %w", err)
	}

	return foundArtefact, nil
}

// parseOrFetchArtefactUUID parses or fetches an artefact uuid based on either a uuid or a / separated compound.
func parseOrFetchArtefactUUID(ctx context.Context, httpClient *http.Client, configuration *Configuration, input string) (uuid.UUID, error) {
	parsedUUID, err := uuid.Parse(input)
	if err == nil {
		return parsedUUID, nil
	}

	inputAsSlice := strings.Split(input, "/")
	if len(inputAsSlice) != 2 {
		return [16]byte{}, fmt.Errorf("input did not match %%s/%%s format: %w", ErrIncorrectArgumentFormat)
	}

	if err != nil {
		return [16]byte{}, fmt.Errorf("failed to find provided artefact %s: %w", input, err)
	}

	return foundArtefact.UUID, nil
}

// parseOrFetchServerUUID parses or fetches a server uuid based on either a uuid or a / separated compound.
func parseOrFetchServerUUID(ctx context.Context, httpClient *http.Client, configuration *Configuration, input string) (uuid.UUID, error) {
	parsedUUID, err := uuid.Parse(input)
	if err == nil {
		return parsedUUID, nil
	}

	inputAsSlice := strings.Split(input, "/")
	if len(inputAsSlice) != 2 {
		return [16]byte{}, fmt.Errorf("input did not match %%s/%%s format: %w", ErrIncorrectArgumentFormat)
	}

	if err != nil {
		return [16]byte{}, fmt.Errorf("failed to find provided server %s: %w", input, err)
	}

	return foundServer.UUID, nil
}
