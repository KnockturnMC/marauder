package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"

	"gitea.knockturnmc.com/marauder/lib/pkg/artefact"

	"gitea.knockturnmc.com/marauder/lib/pkg/models/networkmodel"
	"gitea.knockturnmc.com/marauder/lib/pkg/utils"

	"github.com/gonvenience/bunt"
	"github.com/spf13/cobra"
)

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

		return publishArtefactInternalExecute(ctx, cmd, httpClient, configuration, artefactFilePath, artefactFileSignaturePath)
	}

	return command
}

// publishArtefactInternalExecute is the internal command execution logic for the publish artefact sub command.
func publishArtefactInternalExecute(
	ctx context.Context,
	cmd *cobra.Command,
	httpClient *http.Client,
	configuration *Configuration,
	artefactFilePath, artefactFileSignaturePath string,
) error {
	// Create multipart writer
	var body bytes.Buffer
	multipartWriter := multipart.NewWriter(&body)

	// Write artefact
	err := utils.WriteFileToMultipart(multipartWriter, artefactFilePath, "artefact")
	if err != nil {
		return fmt.Errorf("failed to write artefact to request body: %w", err)
	}

	// Write signature
	err = utils.WriteFileToMultipart(multipartWriter, artefactFileSignaturePath, "signature")
	if err != nil {
		return fmt.Errorf("failed to write signature to request body: %w", err)
	}

	// Close the writer to finalise writing and flush to the bytes.Buffer
	if err := multipartWriter.Close(); err != nil {
		return fmt.Errorf("failed to close multipart writer: %w", err)
	}

	// post the to controller
	uploadEndpoint := fmt.Sprintf("%s/artefact", configuration.ControllerHost)

	cmd.PrintErrln(bunt.Sprintf("Gray{uploading artefact to %s}", uploadEndpoint))
	response, err := utils.PerformHTTPRequest(ctx, httpClient, http.MethodPost, uploadEndpoint, multipartWriter.FormDataContentType(), &body)
	if err != nil {
		return fmt.Errorf("failed to post to controller: %w", err)
	}

	defer func() { _ = response.Body.Close() }()

	bodyBytes, _ := io.ReadAll(response.Body)
	if !utils.IsOkayStatusCode(response.StatusCode) {
		if response.StatusCode != http.StatusConflict {
			cmd.Println(bunt.Sprintf("Red{failed to upload artefact, controller error: %s}", string(bodyBytes)))
			return nil
		}

		return nil // TODO check if to-be-published artefact matches existing artefact.
	}

	var artefactResult networkmodel.ArtefactModel
	if err := json.Unmarshal(bodyBytes, &artefactResult); err != nil {
		return fmt.Errorf("failed to unmarshal controller result %s: %w", string(bodyBytes), err)
	}

	cmd.PrintErrln(bunt.Sprintf("LimeGreen{successfully uploaded artefact to controller}"))
	cmd.SetContext(context.WithValue(cmd.Context(), KeyPublishResultArtefactModel, artefactResult)) //nolint:contextcheck
	cmd.Println(string(bodyBytes))

	return nil
}

// publishArtefactCheckExisting is responsible for checking if the artefact at the given path exists on the controller.
func publishArtefactCheckExisting(ctx *context.Context, command *cobra.Command, httpClient *http.Client, path string) error {
	fileHandle, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open publishing artefact: %w", err)
	}

	defer func() { _ = fileHandle.Close() }()

	_, _ = artefact.ReadManifestFromTarball(fileHandle)

	return nil // TODO work
}
