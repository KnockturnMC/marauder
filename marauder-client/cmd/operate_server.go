package cmd

import (
	"context"
	"fmt"
	"gitea.knockturnmc.com/marauder/lib/pkg/models/networkmodel"
	"gitea.knockturnmc.com/marauder/lib/pkg/utils"
	"github.com/gonvenience/bunt"
	"github.com/spf13/cobra"
	"io"
)

// OperateServerCommand constructs the operate server subcommand.
func OperateServerCommand(
	ctx context.Context,
	config *Configuration,
) *cobra.Command {
	command := &cobra.Command{
		Use:   "server serverRef action",
		Short: "Executes an operation on a given server",
		Args:  cobra.ExactArgs(2),
	}

	command.RunE = func(cmd *cobra.Command, args []string) error {
		actionType := networkmodel.LifecycleChangeActionType(args[1])
		if !networkmodel.KnownLifecycleChangeActionType(actionType) {
			return fmt.Errorf("unknow action %s: %w", actionType, ErrIncorrectArgumentFormat)
		}

		client, err := config.CreateTLSReadyHTTPClient()
		if err != nil {
			cmd.PrintErrln(bunt.Sprintf("#c43f43{failed to enable tls: %s}", err))
		}

		serverUUID, err := client.ResolveServerReference(ctx, args[0])
		if err != nil {
			return fmt.Errorf("failed to fetch server uuid: %w", err)
		}

		err := client.PostToOperator(ctx, serverUUID, actionType)
		if err != nil {
			return fmt.Errorf("failed to perform http request to controller: %w", err)
		}

		if !utils.IsOkayStatusCode(response.StatusCode) {
			faultyResponseBody, err := io.ReadAll(response.Body)
			if err != nil {
				return fmt.Errorf("failed to read not-okay controller response: %w", err)
			}

			cmd.PrintErrln(bunt.Sprintf("Red{failed to apply action %s on %s: %s}", actionType, serverUUID, faultyResponseBody))
			return nil
		}

		cmd.PrintErrln(bunt.Sprintf("LimeGreen{performed action %s on %s}", actionType, serverUUID))
		return nil
	}

	return command
}
