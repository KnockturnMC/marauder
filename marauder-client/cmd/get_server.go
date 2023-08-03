package cmd

import (
	"context"
	"fmt"

	"gitea.knockturnmc.com/marauder/lib/pkg/models/networkmodel"
	"gitea.knockturnmc.com/marauder/lib/pkg/utils"
	"github.com/gonvenience/bunt"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

// GetServerCommand constructs the servers fetch subcommand.
func GetServerCommand(
	ctx context.Context,
	configuration *Configuration,
) *cobra.Command {
	command := &cobra.Command{
		Use:   "server uuid|(environment [name])",
		Short: "Fetch information about servers from the controller",
		Args:  cobra.RangeArgs(1, 2),
	}

	command.RunE = func(cmd *cobra.Command, args []string) error {
		client, err := configuration.CreateTLSReadyHTTPClient()
		if err != nil {
			cmd.PrintErrln(bunt.Sprintf("#c43f43{failed to enable tls: %s}", err))
		}

		resultSlice := make([]networkmodel.ServerModel, 0)

		defer func() { printFetchResult(cmd, resultSlice) }()

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

		// Fetching via environment and identifier.
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
