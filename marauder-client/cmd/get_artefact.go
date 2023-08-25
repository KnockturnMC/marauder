package cmd

import (
	"context"
	"fmt"

	"gitea.knockturnmc.com/marauder/lib/pkg/models/networkmodel"
	"github.com/gonvenience/bunt"
	"github.com/spf13/cobra"
)

// GetArtefactCommand constructs the artefact fetch subcommand.
func GetArtefactCommand(
	ctx context.Context,
	config *Configuration,
) *cobra.Command {
	command := &cobra.Command{
		Use:   "artefact [identifier|reference]",
		Short: "Fetch information about artefacts from the controller",
		Args:  cobra.ExactArgs(1),
	}

	command.RunE = func(cmd *cobra.Command, args []string) error {
		client, err := config.CreateTLSReadyHTTPClient()
		if err != nil {
			cmd.PrintErrln(bunt.Sprintf("#c43f43{failed to enable tls: %s}", err))
		}

		resultSlice := make([]networkmodel.ArtefactModel, 0)

		defer func() { printFetchResult(cmd, resultSlice) }()

		// Attempt to parse uuid.
		artefactUUID, err := client.ResolveArtefactReference(ctx, args[0])
		if err == nil {
			cmd.PrintErrln(bunt.Sprintf("Gray{requesting single artefact %s}", artefactUUID))

			resultStruct, err := client.FetchArtefact(ctx, artefactUUID)
			if err != nil {
				return fmt.Errorf("failed to fetch specific artefact %s: %w", artefactUUID, err)
			}

			resultSlice = append(resultSlice, resultStruct)

			return nil
		}

		cmd.PrintErrln(bunt.Sprintf("Gray{requesting artefacts by identifier}"))

		artefacts, err := client.FetchArtefacts(ctx, args[0])
		if err != nil {
			return fmt.Errorf("failed to fetch artefacts %s: %w", args[0], err)
		}

		resultSlice = artefacts

		return nil
	}

	return command
}
