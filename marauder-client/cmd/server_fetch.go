package cmd

import (
	"context"
	"encoding/json"
	"fmt"

	"gitea.knockturnmc.com/marauder/lib/pkg/models/networkmodel"
	"gitea.knockturnmc.com/marauder/lib/pkg/utils"
	"github.com/gonvenience/bunt"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

// ServersFetchCommand constructs the servers fetch subcommand.
func ServersFetchCommand(
	ctx context.Context,
	configuration *Configuration,
) *cobra.Command {
	command := &cobra.Command{
		Use:   "fetch",
		Short: "Fetch information about servers from the controller",
		Args:  cobra.RangeArgs(1, 2),
	}

	command.RunE = func(cmd *cobra.Command, args []string) error {
		client, err := configuration.CreateTLSReadyHTTPClient()
		if err != nil {
			cmd.PrintErrln(bunt.Sprintf("#c43f43{failed to enable tls: %s}", err))
		}

		resultSlice := make([]networkmodel.ServerModel, 0)

		defer func() { printServerFetchResult(cmd, resultSlice) }()

		// Attempt to parse uuid.
		serverUUID, err := uuid.Parse(args[0])
		if err == nil {
			cmd.PrintErrln(bunt.Sprintf("Gray{requesting server by uuid %s}", serverUUID))

			url := fmt.Sprintf("%s/server/%s", configuration.ControllerHost, serverUUID)
			resultStruct, err := utils.HTTPGetAndBind(ctx, client, url, networkmodel.ServerModel{})
			if err != nil {
				return fmt.Errorf("failed to fetch specific artefact %s: %w", serverUUID, err)
			}

			resultSlice = append(resultSlice, resultStruct)

			return nil
		}

		// Fetching via identifier and version.
		if len(args) == 2 {
			cmd.PrintErrln(bunt.Sprintf("Gray{requesting server by environment and identifier}"))

			url := fmt.Sprintf("%s/servers/%s/%s", configuration.ControllerHost, args[0], args[1])
			servers, err := utils.HTTPGetAndBind(ctx, client, url, networkmodel.ServerModel{})
			if err != nil {
				return fmt.Errorf("failed to fetch servers %s:%s: %w", args[0], args[1], err)
			}

			resultSlice = append(resultSlice, servers)

			return nil
		}

		cmd.PrintErrln(bunt.Sprintf("Gray{requesting servers by environment}"))

		servers, err := utils.HTTPGetAndBind(ctx, client, configuration.ControllerHost+"/servers/"+args[0], resultSlice)
		if err != nil {
			return fmt.Errorf("failed to fetch servers %s: %w", args[0], err)
		}

		resultSlice = servers

		return nil
	}

	return command
}

// printArtefactFetchResult prints the passed result set to the command output stream.
func printServerFetchResult(cmd *cobra.Command, result []networkmodel.ServerModel) {
	output, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		cmd.PrintErrln(bunt.Sprintf("Red{failed to marshal result %s}", err))
		return
	}

	cmd.Println(string(output))
}
