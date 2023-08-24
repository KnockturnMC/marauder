package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// WorkflowBuildAndDeployCommand constructs the workflow build and deploy subcommand.
func WorkflowBuildAndDeployCommand(
	ctx context.Context,
	configuration *Configuration,
) *cobra.Command {
	var manifestFileLocation string

	command := &cobra.Command{
		Use:   "build-and-deploy",
		Short: "Builds, signs and pushes and deploys the local artefact found in the working directory",
		Args:  cobra.MaximumNArgs(1),
	}

	command.PersistentFlags().StringVarP(&manifestFileLocation, "manifest", "m", ".marauder.json", "location of the manifest file")

	command.RunE = func(cmd *cobra.Command, args []string) error {
		workingDirectory := "."
		if len(args) > 0 {
			workingDirectory = args[0]
		}

		defer func() { _ = os.Remove("output.tar") }()

		//nolint:contextcheck
		if err := buildArtefactInternalExecute(cmd, configuration, manifestFileLocation, "output.tar", workingDirectory, true); err != nil {
			return fmt.Errorf("failed to build and sign artefact: %w", err)
		}

		tarballLocation, ok := cmd.Context().Value(KeyBuildCommandTarballOutput).(TarballBuildResult)
		if !ok {
			return fmt.Errorf("failed to retrieve tarball location from build logic %v: %w", tarballLocation, ErrContextMissingValue)
		}

		defer func() { _ = os.Remove(tarballLocation.TarballSignatureLocation) }()

		if err := publishArtefactInternalExecute(
			ctx,
			cmd,
			configuration,
			tarballLocation.TarballFileLocation,
			tarballLocation.TarballSignatureLocation,
		); err != nil {
			return fmt.Errorf("failed to publish artefact to controller: %w", err)
		}

		return nil
	}

	return command
}
