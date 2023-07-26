package worker

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"sync"
)

// ErrBadStatusCode is returned if the controller returned a bad status code.
var ErrBadStatusCode = errors.New("bad status code")

func IsOkayStatusCode(code int) bool {
	return code >= 200 && code <= 400
}

// DownloadService is a threadsafe worker capable of properly handling download requests
// from multiple threads at the same time.
type DownloadService interface {
	// Download downloads the file from the given url into a service specific cache folder under the passed name.
	// The method returns the full path of the file, ready for future use or an error.
	Download(ctx context.Context, url string, filename string) (string, error)
}

// DownloadResult is a smaller helper utility representing the result retrieved from a download dispatcher.
type DownloadResult struct {
	fullFileName string
}

type MutexDownloadService struct {
	httpClient           *http.Client
	dispatcher           *Dispatcher[DownloadResult]
	resultListeners      map[string][]chan Outcome[DownloadResult]
	resultListenersMutex *sync.Mutex
	cacheDirectory       string
}

// NewMutexDownloadService creates a new download service using a mutex.
func NewMutexDownloadService(httpClient *http.Client, dispatcher *Dispatcher[DownloadResult], cacheDirectory string) *MutexDownloadService {
	return &MutexDownloadService{
		httpClient:           httpClient,
		dispatcher:           dispatcher,
		resultListeners:      make(map[string][]chan Outcome[DownloadResult]),
		resultListenersMutex: &sync.Mutex{},
		cacheDirectory:       cacheDirectory,
	}
}

func (m *MutexDownloadService) Download(ctx context.Context, url string, filename string) (string, error) {
	m.resultListenersMutex.Lock() // Lock here, we are going to read from the map.

	listener, ok := m.resultListeners[url]
	resultChan := make(chan Outcome[DownloadResult])
	if ok {
		m.resultListeners[url] = append(listener, resultChan)
	} else {
		m.resultListeners[url] = []chan Outcome[DownloadResult]{resultChan}
		dispatchResult := m.dispatcher.Dispatch(func() (DownloadResult, error) {
			file, err := m.downloadURLToFile(ctx, url, filename)
			if err != nil {
				return DownloadResult{}, fmt.Errorf("failed to download url to file: %w", err)
			}

			return DownloadResult{
				fullFileName: file,
			}, nil
		})

		go func() {
			outcome := <-dispatchResult
			m.resultListenersMutex.Lock()
			defer m.resultListenersMutex.Unlock()
			for _, listener := range m.resultListeners[url] {
				listener <- outcome
			}

			delete(m.resultListeners, url)
		}()
	}

	m.resultListenersMutex.Unlock()

	outcome := <-resultChan
	if outcome.Err != nil {
		return "", fmt.Errorf("failed to download file %s: %w", url, outcome.Err)
	}

	return outcome.Value.fullFileName, nil
}

func (m *MutexDownloadService) downloadURLToFile(ctx context.Context, url string, filename string) (string, error) {
	downloadTargetPath := path.Join(m.cacheDirectory, path.Base(filename))

	downloadReq, err := http.NewRequestWithContext(ctx, http.MethodGet, url, &bytes.Buffer{})
	if err != nil {
		return "", fmt.Errorf("failed to create download request: %w", err)
	}

	downloadResponse, err := m.httpClient.Do(downloadReq)
	if err != nil {
		return "", fmt.Errorf("failed to execute download request: %w", err)
	}

	defer func() { _ = downloadResponse.Body.Close() }()

	if !IsOkayStatusCode(downloadResponse.StatusCode) {
		return "", fmt.Errorf("failed to download %s, status code %d: %w", url, downloadResponse.StatusCode, ErrBadStatusCode)
	}

	downloadTarget, err := os.Create(downloadTargetPath)
	if err != nil {
		return "", fmt.Errorf("failed to open output file: %w", err)
	}

	defer func() { _ = downloadTarget.Close() }()

	if _, err := io.Copy(downloadTarget, downloadResponse.Body); err != nil {
		return "", fmt.Errorf("failed to copy download response to file system: %w", err)
	}

	return downloadTargetPath, nil
}
