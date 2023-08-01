package cmd

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"

	"github.com/gonvenience/bunt"
	"github.com/spf13/cobra"
)

// ArtefactPublishCommand constructs the artefact publish subcommand.
func ArtefactPublishCommand(
	ctx context.Context,
	configuration *Configuration,
) *cobra.Command {
	command := &cobra.Command{
		Use:   "publish",
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

		// Create multipart writer
		var body bytes.Buffer
		multipartWriter := multipart.NewWriter(&body)

		// Write artefact
		err = writeFileToMultipart(multipartWriter, artefactFilePath, "artefact")
		if err != nil {
			return fmt.Errorf("failed to write artefact to request body: %w", err)
		}

		// Write signature
		err = writeFileToMultipart(multipartWriter, artefactFileSignaturePath, "signature")
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
		response, err := doPost(ctx, httpClient, uploadEndpoint, multipartWriter.FormDataContentType(), body)
		if err != nil {
			return fmt.Errorf("failed to post to controller: %w", err)
		}

		defer func() { _ = response.Body.Close() }()

		bodyBytes, _ := io.ReadAll(response.Body)
		if response.StatusCode >= http.StatusBadRequest || response.StatusCode < http.StatusOK {
			cmd.Println(bunt.Sprintf("Red{failed to upload artefact, controller error: %s}", string(bodyBytes)))

			return nil
		}

		cmd.PrintErrln(bunt.Sprintf("LimeGreen{successfully uploaded artefact to controller}"))

		cmd.Println(string(bodyBytes))

		return nil
	}

	return command
}

// doPost creates a post request and publishes it to the passed http client.
func doPost(ctx context.Context, httpClient *http.Client, controllerHost string, contentType string, body bytes.Buffer) (*http.Response, error) {
	postRequest, err := http.NewRequestWithContext(ctx, http.MethodPost, controllerHost, &body)
	if err != nil {
		return nil, fmt.Errorf("failed to create post request: %w", err)
	}

	postRequest.Header.Set("Content-Type", contentType)

	response, err := httpClient.Do(postRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to execute post request: %w", err)
	}

	return response, nil
}

// writeFileToMultipart writes the passed file to the multipart writer.
func writeFileToMultipart(multipartWriter *multipart.Writer, filePath string, name string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file to upload: %w", err)
	}

	defer func() { _ = file.Close() }()

	artefactSigUpload, err := multipartWriter.CreateFormFile(name, name)
	if err != nil {
		return fmt.Errorf("failed to create form file for %s: %w", name, err)
	}

	if _, err := io.Copy(artefactSigUpload, file); err != nil {
		return fmt.Errorf("failed to write %s to form header: %w", name, err)
	}

	return nil
}
