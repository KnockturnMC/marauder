package cmd

import (
	"context"
	"fmt"

	"gitea.knockturnmc.com/marauder/lib/pkg/models/networkmodel"
	"github.com/gonvenience/bunt"
	"github.com/spf13/cobra"
)

// GetServerCommand constructs the servers fetch subcommand.
func GetServerCommand(
	ctx context.Context,
	config *Configuration,
) *cobra.Command {
	command := &cobra.Command{
		Use:   "server [environment|reference]",
		Short: "Fetch information about servers from the controller",
		Args:  cobra.ExactArgs(1),
	}

	command.RunE = func(cmd *cobra.Command, args []string) error {
		client, err := config.CreateTLSReadyHTTPClient()
		if err != nil {
			cmd.PrintErrln(bunt.Sprintf("#c43f43{failed to enable tls: %s}", err))
		}

		resultSlice := make([]networkmodel.ServerModel, 0)

		defer func() { printFetchResult(cmd, resultSlice) }()

		// Attempt to parse uuid.
		serverUUID, err := client.ResolveServerReference(ctx, args[0])
		if err == nil {
			cmd.PrintErrln(bunt.Sprintf("Gray{requesting server by uuid %s}", serverUUID))

			resultStruct, err := client.FetchServer(ctx, serverUUID)
			if err != nil {
				return fmt.Errorf("failed to fetch specific artefact %s: %w", serverUUID, err)
			}

			resultSlice = append(resultSlice, resultStruct)

			return nil
		}

		cmd.PrintErrln(bunt.Sprintf("Gray{requesting servers by environment}"))

		servers, err := client.FetchServers(ctx, args[0])
		if err != nil {
			return fmt.Errorf("failed to fetch servers %s: %w", args[0], err)
		}

		resultSlice = servers

		return nil
	}

	return command
}
