package cmd

import (
	"context"
	"fmt"

	"gitea.knockturnmc.com/marauder/lib/pkg/models/networkmodel"
	"gitea.knockturnmc.com/marauder/lib/pkg/utils"
	"github.com/gonvenience/bunt"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

// GetArtefactCommand constructs the artefact fetch subcommand.
func GetArtefactCommand(
	ctx context.Context,
	configuration *Configuration,
) *cobra.Command {
	command := &cobra.Command{
		Use:   "artefact <uuid|identifier> [version]",
		Short: "Fetch information about artefacts from the controller",
		Args:  cobra.RangeArgs(1, 2),
	}

	command.RunE = func(cmd *cobra.Command, args []string) error {
		client, err := configuration.CreateTLSReadyHTTPClient()
		if err != nil {
			cmd.PrintErrln(bunt.Sprintf("#c43f43{failed to enable tls: %s}", err))
		}

		resultSlice := make([]networkmodel.ArtefactModel, 0)

		defer func() { printFetchResult(cmd, resultSlice) }()

		// Attempt to parse uuid.
		artefactUUID, err := uuid.Parse(args[0])
		if err == nil {
			cmd.PrintErrln(bunt.Sprintf("Gray{requesting artefact by uuid %s}", artefactUUID))

			url := fmt.Sprintf("%s/artefact/%s", configuration.ControllerHost, artefactUUID)
			resultStruct, err := utils.HTTPGetAndBind(ctx, client, url, networkmodel.ArtefactModel{})
			if err != nil {
				return fmt.Errorf("failed to fetch specific artefact %s: %w", artefactUUID, err)
			}

			resultSlice = append(resultSlice, resultStruct)

			return nil
		}

		// Fetching via identifier and version.
		if len(args) == 2 {
			cmd.PrintErrln(bunt.Sprintf("Gray{requesting artefact by identifier and version}"))

			url := fmt.Sprintf("%s/artefacts/%s/%s", configuration.ControllerHost, args[0], args[1])
			artefact, err := utils.HTTPGetAndBind(ctx, client, url, networkmodel.ArtefactModel{})
			if err != nil {
				return fmt.Errorf("failed to fetch artefact %s:%s: %w", args[0], args[1], err)
			}

			resultSlice = append(resultSlice, artefact)

			return nil
		}

		cmd.PrintErrln(bunt.Sprintf("Gray{requesting artefacts by identifier}"))

		artefacts, err := utils.HTTPGetAndBind(ctx, client, configuration.ControllerHost+"/artefacts/"+args[0], resultSlice)
		if err != nil {
			return fmt.Errorf("failed to fetch artefacts %s: %w", args[0], err)
		}

		resultSlice = artefacts

		return nil
	}

	return command
}
