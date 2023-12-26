package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"reflect"

	"gitea.knockturnmc.com/marauder/lib/pkg/utils"
	"github.com/samber/mo"

	"gitea.knockturnmc.com/marauder/lib/pkg/controller"
	"gitea.knockturnmc.com/marauder/lib/pkg/models/networkmodel"

	"gitea.knockturnmc.com/marauder/lib/pkg/artefact"

	"github.com/gonvenience/bunt"
	"github.com/spf13/cobra"
)

// ErrRemoteManifestMismatch is returned when comparing a remote manifest with a local one and they do not match.
var ErrRemoteManifestMismatch = errors.New("remote manifest mismatch")

// PublishArtefactCommand constructs the artefact publish subcommand.
func PublishArtefactCommand(
	ctx context.Context,
	configuration *Configuration,
) *cobra.Command {
	command := &cobra.Command{
		Use:   "artefact artefactFileName [artefactFileSignatureName]",
		Short: "Publishes a marauder artefact to a controller",
		Args:  cobra.RangeArgs(1, 2),
	}

	command.RunE = func(cmd *cobra.Command, args []string) error {
		// Parse the paths to the files to publish
		artefactFilePath := args[0]
		artefactFileSignaturePath := artefactFilePath + ".sig"
		if len(args) > 1 {
			artefactFileSignaturePath = args[1]
		}

		// create http client
		httpClient, err := configuration.CreateTLSReadyHTTPClient()
		if err != nil {
			cmd.PrintErrln(bunt.Sprintf("#c43f43{failed to enable tls: %s}", err))
		}

		return publishArtefactInternalExecute(ctx, cmd, httpClient, artefactFilePath, artefactFileSignaturePath)
	}

	return command
}

// publishArtefactInternalExecute is the internal command execution logic for the publish artefact sub command.
func publishArtefactInternalExecute(
	ctx context.Context,
	cmd *cobra.Command,
	client controller.Client,
	artefactFilePath, signatureFilePath string,
) error {
	artefactFileHandle, err := os.Open(artefactFilePath)
	if err != nil {
		return fmt.Errorf("failed to open artefact file: %w", err)
	}

	defer func() { _ = artefactFileHandle.Close() }()

	signatureFileHandle, err := os.Open(signatureFilePath)
	if err != nil {
		return fmt.Errorf("failed to open artefact file: %w", err)
	}

	defer func() { _ = signatureFileHandle.Close() }()

	publishArtefact, statusCode, err := client.PublishArtefact(
		ctx,
		artefactFileHandle,
		signatureFileHandle,
	)

	if err != nil && statusCode.IsAbsent() {
		return fmt.Errorf("failed to publish: %w", err)
	}

	// publishing did not work out, attempt to find duplicate
	if err != nil {
		remoteArtefact, err := publishArtefactTryRecoverConflict(ctx, client, statusCode, artefactFileHandle)
		if err != nil {
			return fmt.Errorf("failed to publish artefact: %w", err)
		}

		cmd.PrintErrln(bunt.Sprintf("Gray{found existing matching artefact on controller}"))
		publishArtefact = remoteArtefact
	}

	cmd.PrintErrln(bunt.Sprintf("LimeGreen{successfully uploaded artefact to controller}"))
	cmd.SetContext(context.WithValue(cmd.Context(), KeyPublishResultArtefactModel, publishArtefact)) //nolint:contextcheck
	publishedArtefactAsJSON, err := json.Marshal(publishArtefact)
	if err != nil {
		cmd.Println(fmt.Sprintf("%v", publishArtefact))
		return fmt.Errorf("failed to marshal published artefact to string: %w", err)
	}

	cmd.Println(string(publishedArtefactAsJSON))

	return nil
}

func publishArtefactTryRecoverConflict(
	ctx context.Context,
	client controller.Client,
	statusCode mo.Option[int],
	artefactFileHandle *os.File,
) (networkmodel.ArtefactModel, error) {
	statusCodeAsInt := statusCode.MustGet()
	if statusCodeAsInt != http.StatusConflict {
		return networkmodel.ArtefactModel{}, fmt.Errorf("failed to publish: %w", utils.ErrStatusCodeUnrecoverable)
	}

	// We have a conflict, the version already exists.
	if _, err := artefactFileHandle.Seek(0, io.SeekStart); err != nil {
		return networkmodel.ArtefactModel{}, fmt.Errorf("failed to reset artefact file handle: %w", err)
	}

	remoteArtefact, err := publishArtefactCheckExisting(ctx, client, artefactFileHandle)
	if errors.Is(err, ErrRemoteManifestMismatch) {
		return networkmodel.ArtefactModel{}, fmt.Errorf("conflict on publish, remote artefact exists but does not match: %w", err)
	} else if err != nil {
		return networkmodel.ArtefactModel{}, fmt.Errorf("failed to compare conflicting remote artefact with local one: %w", err)
	}

	return remoteArtefact, nil
}

// publishArtefactCheckExisting is responsible for checking if the artefact at the given path exists on the controller.
func publishArtefactCheckExisting(ctx context.Context, client controller.Client, artefactFileHandle io.Reader) (networkmodel.ArtefactModel, error) {
	manifest, err := artefact.ReadManifestFromTarball(artefactFileHandle)
	if err != nil {
		return networkmodel.ArtefactModel{}, fmt.Errorf("failed to find manifest: %w", err)
	}

	remoteArtefact, err := client.FetchArtefactByIdentifierAndVersion(ctx, manifest.Identifier, manifest.Version)
	if err != nil {
		return networkmodel.ArtefactModel{}, fmt.Errorf("failed to find remote artefact: %w", err)
	}

	remoteArtefactManifest, err := client.FetchManifest(ctx, remoteArtefact.UUID)
	if err != nil {
		return networkmodel.ArtefactModel{}, fmt.Errorf("failed to fetch remote artefacts manifest: %w", err)
	}

	if !reflect.DeepEqual(manifest.Files, remoteArtefactManifest.Files) {
		return networkmodel.ArtefactModel{}, fmt.Errorf("manifests do not match completely: %w", ErrRemoteManifestMismatch)
	}

	return remoteArtefact, nil
}
