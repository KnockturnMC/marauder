package utils

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

// ErrBadStatusCode is returned if the controller returned a bad status code.
var ErrBadStatusCode = errors.New("bad status code")

// HTTPGetAndBind performs a get request using the http client at the given path and binds the result into
// the passed struct.
// If a response code that is not 200<=code<=400, an error is returned.
func HTTPGetAndBind[T any](ctx context.Context, client *http.Client, path string, bindTarget T) (T, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, path, &bytes.Buffer{})
	if err != nil {
		return bindTarget, fmt.Errorf("failed to create http request: %w", err)
	}

	resp, err := client.Do(request)
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
	return code >= 200 && code <= 400
}
