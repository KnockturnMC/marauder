package cmd

import (
	"context"
	"fmt"

	"github.com/gonvenience/bunt"
	"github.com/knockturnmc/marauder/marauder-proto/src/main/golang/marauderpb"
	"github.com/spf13/cobra"
)

// ManageServerPlayersCommand constructs the servers fetch subcommand.
func ManageServerPlayersCommand(
	ctx context.Context,
	config *Configuration,
) *cobra.Command {
	command := &cobra.Command{
		Use:   "players serverUUID",
		Short: "Fetches the live player list from the specified server uuid",
		Args:  cobra.RangeArgs(1, 2),
	}

	command.RunE = func(cmd *cobra.Command, args []string) error {
		client, err := config.CreateTLSReadyHTTPClient()
		if err != nil {
			cmd.PrintErrln(bunt.Sprintf("#c43f43{failed to enable tls: %s}", err))
		}

		resultSlice := make([]marauderpb.Player, 0)

		defer func() { printFetchResult(cmd, resultSlice) }()

		// Attempt to parse uuid.
		serverUUID, err := client.ResolveServerReference(ctx, args[0])
		if err != nil {
			return fmt.Errorf("failed to fetch server uuid: %w", err)
		}

		cmd.PrintErrln(bunt.Sprintf("Gray{requesting players for %s}", serverUUID))

		artefacts, err := client.FetchServerPlayers(ctx, serverUUID)
		if err != nil {
			return fmt.Errorf("failed to fetch server state for %s: %w", args[0], err)
		}

		resultSlice = append(resultSlice, artefacts...)

		return nil
	}

	return command
}
