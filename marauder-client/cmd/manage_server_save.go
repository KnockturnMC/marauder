package cmd

import (
	"context"
	"fmt"
	"strconv"

	"github.com/gonvenience/bunt"
	"github.com/spf13/cobra"
)

// ManageServerToggleSaveCommand( creates the save command for management.
func ManageServerToggleSaveCommand(
	ctx context.Context,
	config *Configuration,
) *cobra.Command {
	command := &cobra.Command{
		Use:   "togglesave flag serverUUID",
		Short: "Toggles if the passed server should be saving files to disk.",
		Args:  cobra.ExactArgs(2),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
			if len(args) == 0 {
				return []cobra.Completion{"true", "false"}, cobra.ShellCompDirectiveNoFileComp
			}
			return nil, cobra.ShellCompDirectiveNoFileComp
		},
	}

	command.RunE = func(cmd *cobra.Command, args []string) error {
		client, err := config.CreateTLSReadyHTTPClient()
		if err != nil {
			cmd.PrintErrln(bunt.Sprintf("#c43f43{failed to enable tls: %s}", err))
		}

		saveFlag, err := strconv.ParseBool(args[0])
		if err != nil {
			return fmt.Errorf("failed to parse save state to bool: %w", err)
		}

		// Attempt to parse uuid.
		serverUUID, err := client.ResolveServerReference(ctx, args[1])
		if err != nil {
			return fmt.Errorf("failed to fetch server uuid: %w", err)
		}

		cmd.PrintErrln(bunt.Sprintf("Gray{requesting players for %s}", serverUUID))

		if err := client.ManageServerToggleSave(ctx, serverUUID, saveFlag); err != nil {
			return fmt.Errorf("failed to toggle server save: %w", err)
		}

		cmd.PrintErrln(bunt.Sprintf("LimeGreen{toggled save for %s to %s}", serverUUID, strconv.FormatBool(saveFlag)))
		return nil
	}

	return command
}
