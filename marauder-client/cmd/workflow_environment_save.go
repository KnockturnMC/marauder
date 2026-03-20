package cmd

import (
	"context"
	"fmt"
	"strconv"

	"github.com/Goldziher/go-utils/sliceutils"
	"github.com/gonvenience/bunt"
	"github.com/knockturnmc/marauder/marauder-lib/pkg/models/networkmodel"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// WorkflowToggleEnvironmentSave constructs the workflow to toggle an entire environments save state.
func WorkflowToggleEnvironmentSave(
	ctx context.Context,
	configuration *Configuration,
) *cobra.Command {
	command := &cobra.Command{
		Use:   "toggle-env-save environment state",
		Short: "Fetches and then toggles the save state for all servers with a management socket in an environment",
		Args:  cobra.ExactArgs(2),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
			if len(args) == 1 {
				return []cobra.Completion{"true", "false"}, cobra.ShellCompDirectiveNoFileComp
			}
			return nil, cobra.ShellCompDirectiveNoFileComp
		},
	}

	command.RunE = func(cmd *cobra.Command, args []string) error {
		client, err := configuration.CreateTLSReadyHTTPClient()
		if err != nil {
			cmd.PrintErrln(bunt.Sprintf("#c43f43{failed to enable tls: %s}", err))
		}

		environment := args[0]
		stateBool, err := strconv.ParseBool(args[1])
		if err != nil {
			return fmt.Errorf("failed to parse save state '%s' to bool: %w", args[1], err)
		}

		servers, err := client.FetchServers(ctx, environment)
		if err != nil {
			return fmt.Errorf("failed to fetch servers for environment '%s': %w", environment, err)
		}

		servers = sliceutils.Filter(servers, func(value networkmodel.ServerModel, index int, slice []networkmodel.ServerModel) bool {
			return value.ManagementSocketPath != ""
		})

		var lastErr error
		for _, server := range servers {
			if err := client.ManageServerToggleSave(ctx, server.UUID, stateBool); err != nil {
				lastErr = fmt.Errorf("failed to toggle save for server '%s': %w", server.UUID, err)
				logrus.Error(lastErr)
				continue
			}

			cmd.PrintErrln(bunt.Sprintf(
				"LimeGreen{toggled save state for %s/%s to %s}", server.Environment, server.Name, strconv.FormatBool(stateBool),
			))
		}

		return lastErr
	}

	return command
}
