package cmd

import (
	"context"
	"fmt"

	"github.com/gonvenience/bunt"
	"github.com/spf13/cobra"
)

// GetArtefactManifestCommand constructs the artefact manifest fetch subcommand.
func GetArtefactManifestCommand(
	ctx context.Context,
	config *Configuration,
) *cobra.Command {
	command := &cobra.Command{
		Use:   "manifest uuid",
		Short: "Fetch the manifest of an artefact from the controller",
		Args:  cobra.ExactArgs(1),
	}

	command.RunE = func(cmd *cobra.Command, args []string) error {
		client, err := config.CreateTLSReadyHTTPClient()
		if err != nil {
			cmd.PrintErrln(bunt.Sprintf("#c43f43{failed to enable tls: %s}", err))
		}

		// Attempt to parse uuid.
		artefactUUID, err := client.ResolveArtefactReference(ctx, args[0])
		if err != nil {
			return fmt.Errorf("failed to fetch artefact uuid: %w", err)
		}

		cmd.PrintErrln(bunt.Sprintf("Gray{requesting manifest by identifier}"))

		manifest, err := client.FetchManifest(ctx, artefactUUID)
		if err != nil {
			return fmt.Errorf("failed to fetch artefacts %s: %w", args[0], err)
		}

		printFetchResult(cmd, manifest)

		return nil
	}

	return command
}
