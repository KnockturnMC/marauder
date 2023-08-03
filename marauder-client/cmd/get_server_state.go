package cmd

import (
	"context"
	"fmt"
	"strings"

	"gitea.knockturnmc.com/marauder/lib/pkg/models/networkmodel"
	"gitea.knockturnmc.com/marauder/lib/pkg/utils"
	"github.com/gonvenience/bunt"
	"github.com/spf13/cobra"
)

// GetServerStateCommand constructs the servers fetch subcommand.
func GetServerStateCommand(
	ctx context.Context,
	config *Configuration,
) *cobra.Command {
	command := &cobra.Command{
		Use:   "state serverUUID [state]",
		Short: "Fetch information about a servers state from the controller",
		Args:  cobra.RangeArgs(1, 2),
	}

	command.RunE = func(cmd *cobra.Command, args []string) error {
		client, err := config.CreateTLSReadyHTTPClient()
		if err != nil {
			cmd.PrintErrln(bunt.Sprintf("#c43f43{failed to enable tls: %s}", err))
		}

		resultSlice := make([]networkmodel.ArtefactModel, 0)

		defer func() { printFetchResult(cmd, resultSlice) }()

		// Attempt to parse uuid.
		serverUUID, err := parseOrFetchServerUUID(ctx, client, config, args[0])
		if err != nil {
			return fmt.Errorf("failed to fetch server uuid: %w", err)
		}

		var stateType networkmodel.ServerStateType = networkmodel.IS
		// Fetching via identifier and version.
		if len(args) == 2 {
			stateType = networkmodel.ServerStateType(strings.ToUpper(args[1]))
			if !networkmodel.KnownServerStateType(stateType) {
				return fmt.Errorf("failed to parse passed state %s: %w", stateType, networkmodel.ErrUnknownServerState)
			}
		}

		cmd.PrintErrln(bunt.Sprintf("Gray{requesting state %s for %s}", stateType, serverUUID))

		servers, err := utils.HTTPGetAndBind(ctx, client, fmt.Sprintf(
			"%s/server/%s/state/%s",
			config.ControllerHost, serverUUID, string(stateType),
		), resultSlice)
		if err != nil {
			return fmt.Errorf("failed to fetch server state for %s: %w", args[0], err)
		}

		resultSlice = servers

		return nil
	}

	return command
}
