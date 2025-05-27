package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/Goldziher/go-utils/sliceutils"
	"github.com/gonvenience/bunt"
	"github.com/knockturnmc/marauder/marauder-lib/pkg/controller"
	"github.com/knockturnmc/marauder/marauder-lib/pkg/models/networkmodel"
	"github.com/spf13/cobra"
)

// OperateServerCommand constructs the operate server subcommand.
func OperateServerCommand(
	ctx context.Context,
	config *Configuration,
) *cobra.Command {
	var delay time.Duration

	command := &cobra.Command{
		Use:   "server action servers...",
		Short: "Executes the operation on the passed servers",
		Args:  cobra.MinimumNArgs(2),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) == 0 {
				return sliceutils.Map(
					networkmodel.KnownLifecycleActions(),
					func(value networkmodel.LifecycleAction, index int, slice []networkmodel.LifecycleAction) string {
						return string(value)
					},
				), cobra.ShellCompDirectiveDefault
			}

			return nil, cobra.ShellCompDirectiveDefault
		},
	}

	command.PersistentFlags().DurationVar(&delay, "delay", 0, "delay before executing a potential restart")

	command.RunE = func(cmd *cobra.Command, args []string) error {
		actionType := networkmodel.LifecycleAction(args[0])
		if !networkmodel.KnownLifecycleChangeActionType(actionType) {
			return fmt.Errorf("unknow action %s: %w", actionType, ErrIncorrectArgumentFormat)
		}

		client, err := config.CreateTLSReadyHTTPClient()
		if err != nil {
			cmd.PrintErrln(bunt.Sprintf("#c43f43{failed to enable tls: %s}", err))
		}

		return operateServerInternalExecute(
			ctx,
			cmd,
			client,
			actionType,
			delay,
			args[1:],
		)
	}

	return command
}

// operateServerInternalExecute is the internal logic that runs the lifecycle actions for the passed servers.
func operateServerInternalExecute(
	ctx context.Context,
	cmd *cobra.Command,
	client controller.Client,
	lifecycleActionType networkmodel.LifecycleAction,
	delay time.Duration,
	serverIdentifiers []string,
) error {
	// Iterate over servers
	var resultingErr error
	for i := range serverIdentifiers {
		serverUUID, err := client.ResolveServerReference(ctx, serverIdentifiers[i])
		if err != nil {
			return fmt.Errorf("failed to fetch server uuid at %d: %w", i, err)
		}

		if err := client.ExecuteActionOn(ctx, serverUUID, lifecycleActionType, delay); err != nil {
			cmd.PrintErrln(bunt.Sprintf("Red{failed to execute lifecycle action on server %s: %s}", serverUUID, err.Error()))
			resultingErr = err
		} else {
			cmd.PrintErrln(bunt.Sprintf("LimeGreen{performed action %s to %s}", lifecycleActionType, serverUUID))
		}
	}

	return resultingErr
}
