package cmd

import (
	"context"
	"fmt"

	"gitea.knockturnmc.com/marauder/lib/pkg/controller"

	"gitea.knockturnmc.com/marauder/lib/pkg/models/networkmodel"
	"github.com/gonvenience/bunt"
	"github.com/spf13/cobra"
)

func DeployArtefactCommand(
	ctx context.Context,
	config *Configuration,
) *cobra.Command {
	command := &cobra.Command{
		Use:   "artefact artefactUUID servers...",
		Short: "Patches a new deployment target onto ",
		Args:  cobra.MinimumNArgs(2),
	}

	command.RunE = func(cmd *cobra.Command, args []string) error {
		client, err := config.CreateTLSReadyHTTPClient()
		if err != nil {
			cmd.PrintErrln(bunt.Sprintf("#c43f43{failed to enable tls: %s}", err))
		}

		artefactUUIDOrIdentifier := args[0]
		artefactUUID, err := client.ResolveArtefactReference(ctx, artefactUUIDOrIdentifier)
		if err != nil {
			return fmt.Errorf("failed to find artefact uuid: %w", err)
		}

		artefact, err := client.FetchArtefact(ctx, artefactUUID)
		if err != nil {
			return fmt.Errorf("failed to fetch artefact information to deploy: %w", err)
		}

		return deployArtefactInternalExecute(ctx, cmd, client, networkmodel.UpdateServerStateRequest{
			ArtefactIdentifier: artefact.Identifier,
			ArtefactUUID:       &artefactUUID,
		}, args[1:])
	}

	return command
}

func deployArtefactInternalExecute(
	ctx context.Context,
	cmd *cobra.Command,
	client controller.Client,
	updateRequest networkmodel.UpdateServerStateRequest,
	serverIdentifiers []string,
) error {
	// Iterate over servers
	var resultingErr error
	for i := range len(serverIdentifiers) {
		serverUUID, err := client.ResolveServerReference(ctx, serverIdentifiers[i])
		if err != nil {
			return fmt.Errorf("failed to fetch server uuid at %d: %w", i, err)
		}

		if err := client.UpdateState(ctx, serverUUID, networkmodel.TARGET, networkmodel.UpdateServerStateRequest{
			ArtefactIdentifier: updateRequest.ArtefactIdentifier,
			ArtefactUUID:       updateRequest.ArtefactUUID,
		}); err != nil {
			cmd.PrintErrln(bunt.Sprintf("Red{failed to patch server %s: %s}", serverUUID, err.Error()))
			resultingErr = err
		} else {
			cmd.PrintErrln(bunt.Sprintf("LimeGreen{deployed to %s}", serverUUID))
		}
	}

	return resultingErr
}
