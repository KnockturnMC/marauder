package utils

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
)

// ErrBadStatusCode is returned if the controller returned a bad status code.
var ErrBadStatusCode = errors.New("bad status code")

// ErrStatusCodeUnrecoverable is returned if a logic that may recover a status code failed.
var ErrStatusCodeUnrecoverable = errors.New("status code unrecoverable")

// HTTPGetAndBind performs a get request using the http client at the given path and binds the result into
// the passed struct.
// If a response code that is not 200<=code<=400, an error is returned.
func HTTPGetAndBind[T any](ctx context.Context, client *http.Client, path string, bindTarget T) (T, error) {
	resp, err := PerformHTTPRequest(ctx, client, http.MethodGet, path, "application/json", &bytes.Buffer{})
	if err != nil {
		return bindTarget, fmt.Errorf("failed to perform get request: %w", err)
	}

	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return bindTarget, fmt.Errorf("failed to read body of get request: %w", err)
	}

	if !IsOkayStatusCode(resp.StatusCode) {
		return bindTarget, fmt.Errorf("remote returned '%s' (%d): %w", string(body), resp.StatusCode, ErrBadStatusCode)
	}

	if err := json.Unmarshal(body, &bindTarget); err != nil {
		return bindTarget, fmt.Errorf("failed to bind response %s to bind target: %w", string(body), err)
	}

	return bindTarget, nil
}

// IsOkayStatusCode defines if a status code is considered okay.
func IsOkayStatusCode(code int) bool {
	return code >= 200 && code < 400
}

// IsOkayStatusCodeOrErrorWithBody checks IsOkayStatusCode on the passed response and, if the status code is not okay
// creates an error that includes the entire body.
// A ErrBadStatusCode is yielded back if the status code is not okay.
func IsOkayStatusCodeOrErrorWithBody(response *http.Response) error {
	if IsOkayStatusCode(response.StatusCode) {
		return nil
	}

	bodyBytes, err := io.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("failed to read body from failed http request: %w", err)
	}

	return fmt.Errorf("got %s: %w", string(bodyBytes), ErrBadStatusCode)
}

// PerformHTTPRequest creates a request and publishes it to the passed http client.
func PerformHTTPRequest(
	ctx context.Context,
	httpClient *http.Client,
	method string,
	url string,
	contentType string,
	body *bytes.Buffer,
) (*http.Response, error) {
	postRequest, err := http.NewRequestWithContext(ctx, method, url, body)
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

// WriteFileToMultipart writes the passed file to the multipart writer.
func WriteFileToMultipart(multipartWriter *multipart.Writer, file io.Reader, name string) error {
	artefactSigUpload, err := multipartWriter.CreateFormFile(name, name)
	if err != nil {
		return fmt.Errorf("failed to create form file for %s: %w", name, err)
	}

	if _, err := io.Copy(artefactSigUpload, file); err != nil {
		return fmt.Errorf("failed to write %s to form header: %w", name, err)
	}

	return nil
}
