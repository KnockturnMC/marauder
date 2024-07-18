package cmd

import (
	"context"
	"fmt"
	"os"

	"gitea.knockturnmc.com/marauder/lib/pkg/utils"
	"github.com/gonvenience/bunt"
	"github.com/spf13/cobra"
)

// GetArtefactArchiveCommand constructs the artefact archive download subcommand.
func GetArtefactArchiveCommand(
	ctx context.Context,
	config *Configuration,
) *cobra.Command {
	command := &cobra.Command{
		Use:   "archive uuid",
		Short: "Fetch the archive of an artefact from the controller",
		Args:  cobra.ExactArgs(1),
	}

	var outputLocation string
	command.PersistentFlags().StringVarP(&outputLocation, "output", "o", "", "defines location to write the archive to")

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

		cmd.PrintErrln(bunt.Sprintf("Gray{requesting archive by identifier}"))

		// Download manifest
		manifest, err := client.FetchManifest(ctx, artefactUUID)
		if err != nil {
			return fmt.Errorf("failed to fetch artefacts %s: %w", args[0], err)
		}

		// Download artefact archive
		file, err := client.DownloadArtefact(ctx, artefactUUID)
		if err != nil {
			return fmt.Errorf("failed to download artefact archive: %w", err)
		}

		actualOutput := outputLocation
		if actualOutput == "" {
			actualOutput = manifest.Identifier + "-" + manifest.Version + ".tar.gz"
		}

		if err := utils.CopyFile(file, actualOutput); err != nil {
			return fmt.Errorf("failed to copy downloaded artefact: %w", err)
		}

		if err := os.Remove(file); err != nil {
			return fmt.Errorf("failed to delete artefact copy: %w", err)
		}

		return nil
	}

	return command
}
