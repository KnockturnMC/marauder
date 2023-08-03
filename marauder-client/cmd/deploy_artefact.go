package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"gitea.knockturnmc.com/marauder/lib/pkg/models/networkmodel"
	"gitea.knockturnmc.com/marauder/lib/pkg/utils"
	"github.com/gonvenience/bunt"
	"github.com/spf13/cobra"
)

func DeployArtefactCommand(
	ctx context.Context,
	config *Configuration,
) *cobra.Command {
	command := &cobra.Command{
		Use:   "artefact artefactIdentifier artefactUUID servers...",
		Short: "Patches a new deployment target onto ",
		Args:  cobra.MinimumNArgs(3),
	}

	command.RunE = func(cmd *cobra.Command, args []string) error {
		client, err := config.CreateTLSReadyHTTPClient()
		if err != nil {
			cmd.PrintErrln(bunt.Sprintf("#c43f43{failed to enable tls: %s}", err))
		}

		artefactIdentifier := args[0]
		artefactUUID, err := parseOrFetchArtefactUUID(ctx, client, config, args[1])
		if err != nil {
			return fmt.Errorf("failed to find artefact uuid: %w", err)
		}

		updateRequest := networkmodel.UpdateServerStateRequest{
			ArtefactIdentifier: artefactIdentifier,
			ArtefactUUID:       artefactUUID,
		}

		updateRequestAsString, err := json.Marshal(updateRequest)
		if err != nil {
			return fmt.Errorf("failed to marshal update request: %w", err)
		}

		// Iterate over servers
		for i := 2; i < len(args); i++ {
			serverUUID, err := parseOrFetchServerUUID(ctx, client, config, args[i])
			if err != nil {
				return fmt.Errorf("failed to fetch server uuid at %d: %w", i, err)
			}

			response, err := do(ctx, client, http.MethodPatch, fmt.Sprintf(
				"%s/server/%s/state/target",
				config.ControllerHost,
				serverUUID.String(),
			), "application/json", bytes.NewBuffer(updateRequestAsString))
			if err != nil {
				return fmt.Errorf("failed to perform patch request to controller for %s: %w", serverUUID, err)
			}

			if utils.IsOkayStatusCode(response.StatusCode) {
				_ = response.Body.Close()
				cmd.PrintErrln(bunt.Sprintf("LimeGreen{deployed to %s}", serverUUID))

				continue
			}

			faultyResponseBody, err := io.ReadAll(response.Body)
			if err != nil {
				_ = response.Body.Close()
				return fmt.Errorf("failed to read not-okay controller response: %w", err)
			}

			_ = response.Body.Close()
			cmd.PrintErrln(bunt.Sprintf("Red{failed to patch server %s: %s}", serverUUID, faultyResponseBody))

			return nil
		}

		return nil
	}

	return command
}
