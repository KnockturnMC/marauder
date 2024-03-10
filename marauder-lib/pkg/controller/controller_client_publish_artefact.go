package controller

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"

	"gitea.knockturnmc.com/marauder/lib/pkg/models/networkmodel"
	"gitea.knockturnmc.com/marauder/lib/pkg/utils"
	"github.com/samber/mo"
)

// PublishArtefact publishes the artefact read from the given readers to the controller.
// The method returns the status code of the response for further usage.
func (h *HTTPClient) PublishArtefact(ctx context.Context, artefact, signature io.Reader) (networkmodel.ArtefactModel, mo.Option[int], error) {
	// Create multipart writer
	var body bytes.Buffer
	multipartWriter := multipart.NewWriter(&body)

	// Write artefact
	err := utils.WriteFileToMultipart(multipartWriter, artefact, "artefact")
	if err != nil {
		return networkmodel.ArtefactModel{}, mo.None[int](), fmt.Errorf("failed to write artefact to request body: %w", err)
	}

	// Write signature
	err = utils.WriteFileToMultipart(multipartWriter, signature, "signature")
	if err != nil {
		return networkmodel.ArtefactModel{}, mo.None[int](), fmt.Errorf("failed to write signature to request body: %w", err)
	}

	// Close the writer to finalise writing and flush to the bytes.Buffer
	if err := multipartWriter.Close(); err != nil {
		return networkmodel.ArtefactModel{}, mo.None[int](), fmt.Errorf("failed to close multipart writer: %w", err)
	}

	response, err := utils.PerformHTTPRequest(
		ctx,
		h.Client,
		http.MethodPost,
		h.ControllerURL+"/artefact",
		multipartWriter.FormDataContentType(),
		&body,
	)
	if err != nil {
		return networkmodel.ArtefactModel{}, mo.None[int](), fmt.Errorf("failed to post to controller: %w", err)
	}

	// Close response body
	defer func() { _ = response.Body.Close() }()

	bodyBytes, _ := io.ReadAll(response.Body)
	if !utils.IsOkayStatusCode(response.StatusCode) {
		return networkmodel.ArtefactModel{}, mo.Some(response.StatusCode), fmt.Errorf(
			"received non-okay status code (%s): %w", string(bodyBytes), err,
		)
	}

	var artefactResult networkmodel.ArtefactModel
	if err := json.Unmarshal(bodyBytes, &artefactResult); err != nil {
		return networkmodel.ArtefactModel{}, mo.Some(response.StatusCode), fmt.Errorf(
			"failed to unmarshal controller result %s: %w", string(bodyBytes), err,
		)
	}

	return artefactResult, mo.Some(response.StatusCode), nil
}
