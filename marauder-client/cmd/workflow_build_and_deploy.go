package cmd

import (
	"context"
	"fmt"
	"os"

	"gitea.knockturnmc.com/marauder/lib/pkg/controller"

	"gitea.knockturnmc.com/marauder/lib/pkg/models/networkmodel"
	"github.com/Goldziher/go-utils/sliceutils"
	"github.com/gonvenience/bunt"

	"github.com/spf13/cobra"
)

// WorkflowBuildAndDeployCommand constructs the workflow build and deploy subcommand.
func WorkflowBuildAndDeployCommand(
	ctx context.Context,
	configuration *Configuration,
) *cobra.Command {
	var (
		manifestFileLocation   string
		deploymentEnvironment  string
		restartAffectedServers bool
	)

	command := &cobra.Command{
		Use:   "build-and-deploy [workdir]",
		Short: "Builds, signs and pushes and deploys the local artefact found in the working directory",
		Args:  cobra.MaximumNArgs(1),
	}

	command.PersistentFlags().StringVarP(&manifestFileLocation, "manifest", "m", ".marauder.json", "location of the manifest file")
	command.PersistentFlags().StringVarP(&deploymentEnvironment, "env", "e", "", "environment to deploy into")
	command.PersistentFlags().BoolVar(&restartAffectedServers, "restart", false, "restart the servers deployed to")

	_ = command.MarkPersistentFlagRequired("env")

	command.RunE = func(cmd *cobra.Command, args []string) error {
		client, err := configuration.CreateTLSReadyHTTPClient()
		if err != nil {
			cmd.PrintErrln(bunt.Sprintf("#c43f43{failed to enable tls: %s}", err))
		}

		workingDirectory := "."
		if len(args) > 0 {
			workingDirectory = args[0]
		}

		defer func() { _ = os.Remove("output.tar") }()

		//nolint:contextcheck
		if err := buildArtefactInternalExecute(cmd, configuration, manifestFileLocation, "output.tar", workingDirectory, true); err != nil {
			return fmt.Errorf("failed to build and sign artefact: %w", err)
		}

		tarballLocation, valueFound := cmd.Context().Value(KeyBuildCommandTarballOutput).(TarballBuildResult)
		if !valueFound {
			return fmt.Errorf("failed to retrieve tarball location from build logic %v: %w", tarballLocation, ErrContextMissingValue)
		}

		// Delete tarball afterwards
		defer func() { _ = os.Remove(tarballLocation.TarballSignatureLocation) }()

		// publish it
		if err := publishArtefactInternalExecute(
			ctx,
			cmd,
			client,
			tarballLocation.TarballFileLocation,
			tarballLocation.TarballSignatureLocation,
		); err != nil {
			return fmt.Errorf("failed to publish artefact to controller: %w", err)
		}

		publishedArtefactModel, valueFound := cmd.Context().Value(KeyPublishResultArtefactModel).(networkmodel.ArtefactModel)
		if !valueFound {
			return fmt.Errorf("failed to retrieve published artefact from build logic %v: %w", publishedArtefactModel, ErrContextMissingValue)
		}

		if err := workflowBuildAndDeployDeployPublishedArtefact(
			ctx,
			cmd,
			client,
			tarballLocation,
			publishedArtefactModel,
			deploymentEnvironment,
			restartAffectedServers,
		); err != nil {
			return fmt.Errorf("failed to deploy: %w", err)
		}

		return nil
	}

	return command
}

// workflowBuildAndDeployDeployPublishedArtefact deploys a now published artefact to the configured servers and potentially restarts them.
func workflowBuildAndDeployDeployPublishedArtefact(
	ctx context.Context,
	cmd *cobra.Command,
	client controller.Client,
	artefact TarballBuildResult,
	remoteArtefact networkmodel.ArtefactModel,
	deploymentEnvironment string,
	restartAffectedServers bool,
) error {
	serverTargets, valueFound := artefact.Manifest.DeploymentTargets[deploymentEnvironment]
	if !valueFound {
		serverTargets = make([]string, 0)
		cmd.PrintErrln(bunt.Sprintf("Gray{no servers found for environment %s}", deploymentEnvironment))
	}

	// Map them to strings for the deployment function
	serverTargets = sliceutils.Map(serverTargets, func(value string, _ int, _ []string) string {
		return deploymentEnvironment + "/" + value
	})

	cmd.PrintErrln(bunt.Sprintf("Gray{deploying to servers: %v}", serverTargets))

	if err := deployArtefactInternalExecute(
		ctx,
		cmd,
		client,
		networkmodel.UpdateServerStateRequest{
			ArtefactIdentifier: remoteArtefact.Identifier,
			ArtefactUUID:       &remoteArtefact.UUID,
		},
		serverTargets,
	); err != nil {
		return fmt.Errorf("failed to deploy artefact to targeted servers: %w", err)
	}

	// Potentially restart affected servers
	if restartAffectedServers {
		if err := operateServerInternalExecute(
			ctx,
			cmd,
			client,
			networkmodel.UpgradeDeployment,
			serverTargets,
		); err != nil {
			return fmt.Errorf("failed to upgrade affected servers: %w", err)
		}
	}
	return nil
}
