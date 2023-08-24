package cmd

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"

	"gitea.knockturnmc.com/marauder/lib/pkg/models/networkmodel"
	"gitea.knockturnmc.com/marauder/lib/pkg/utils"
	"github.com/gonvenience/bunt"
	"github.com/spf13/cobra"
)

// OperateServerCommand constructs the operate server subcommand.
func OperateServerCommand(
	ctx context.Context,
	config *Configuration,
) *cobra.Command {
	command := &cobra.Command{
		Use:   "server uuid action",
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

		serverUUID, err := parseOrFetchServerUUID(ctx, client, config, args[0])
		if err != nil {
			return fmt.Errorf("failed to fetch server uuid: %w", err)
		}

		response, err := do(ctx, client, http.MethodPost, fmt.Sprintf(
			"%s/operator/%s/server/%s/%s",
			config.ControllerHost, serverUUID, serverUUID, actionType,
		), "plain/text", &bytes.Buffer{})
		if err != nil {
			return fmt.Errorf("failed to perform http request to controller: %w", err)
		}

		defer func() { _ = response.Body.Close() }()

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
