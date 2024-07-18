package worker

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"

	"gitea.knockturnmc.com/marauder/lib/pkg/utils"
)

// DownloadURLTo downloads the contents at the specific url to the passed path.
// The folder containing the path has to exist.
func DownloadURLTo(ctx context.Context, httpClient *http.Client, url string, path string) error {
	downloadReq, err := http.NewRequestWithContext(ctx, http.MethodGet, url, &bytes.Buffer{})
	if err != nil {
		return fmt.Errorf("failed to create download request: %w", err)
	}

	downloadResponse, err := httpClient.Do(downloadReq)
	if err != nil {
		return fmt.Errorf("failed to execute download request: %w", err)
	}

	defer func() { _ = downloadResponse.Body.Close() }()

	if !utils.IsOkayStatusCode(downloadResponse.StatusCode) {
		return fmt.Errorf("failed to download %s, status code %d: %w", url, downloadResponse.StatusCode, utils.ErrBadStatusCode)
	}

	downloadTarget, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to open output file: %w", err)
	}

	defer func() { _ = downloadTarget.Close() }()

	if _, err := io.Copy(downloadTarget, downloadResponse.Body); err != nil {
		return fmt.Errorf("failed to copy download response to file system: %w", err)
	}

	return nil
}
