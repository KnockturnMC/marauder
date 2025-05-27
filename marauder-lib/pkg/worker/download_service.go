package worker

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"sync"
	"time"

	"github.com/Goldziher/go-utils/maputils"
	"github.com/Goldziher/go-utils/sliceutils"
)

// DownloadService is a threadsafe worker capable of properly handling download requests
// from multiple threads at the same time.
type DownloadService interface {
	// Download downloads the file from the given url into a service specific cache folder under the passed name.
	// The method returns the full path of the file, ready for future use or an error.
	Download(ctx context.Context, url string, filename string) (string, error)

	// CleanLocalCache cleans the local download cache.
	// For this, any file older than the provided file age is removed if the download service does not currently have a running job to download said
	// file.
	CleanLocalCache(fileAge time.Duration) error
}

// DownloadResult is a smaller helper utility representing the result retrieved from a download dispatcher.
type DownloadResult struct {
	fullFileName string
}

type RunningDownload struct {
	Listeners      []chan Outcome[DownloadResult]
	TargetFileName string
}

type MutexDownloadService struct {
	httpClient           *http.Client
	dispatcher           *Dispatcher[DownloadResult]
	runningDownloads     map[string]*RunningDownload
	resultListenersMutex *sync.Mutex
	cacheDirectory       string
}

// NewMutexDownloadService creates a new download service using a mutex.
func NewMutexDownloadService(httpClient *http.Client, dispatcher *Dispatcher[DownloadResult], cacheDirectory string) *MutexDownloadService {
	return &MutexDownloadService{
		httpClient:           httpClient,
		dispatcher:           dispatcher,
		runningDownloads:     make(map[string]*RunningDownload),
		resultListenersMutex: &sync.Mutex{},
		cacheDirectory:       cacheDirectory,
	}
}

func (m *MutexDownloadService) Download(ctx context.Context, url string, filename string) (string, error) {
	m.resultListenersMutex.Lock() // Lock here, we are going to read from the map.

	runningDownload, ok := m.runningDownloads[url]
	resultChan := make(chan Outcome[DownloadResult])
	if ok {
		//nolint:gocritic // this is fine, we mutate on purpose
		m.runningDownloads[url].Listeners = append(runningDownload.Listeners, resultChan)
	} else {
		m.runningDownloads[url] = &RunningDownload{
			Listeners:      []chan Outcome[DownloadResult]{resultChan},
			TargetFileName: filename,
		}

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
			for _, listener := range m.runningDownloads[url].Listeners {
				listener <- outcome
			}

			delete(m.runningDownloads, url)
		}()
	}

	m.resultListenersMutex.Unlock()

	outcome := <-resultChan
	if outcome.Err != nil {
		return "", fmt.Errorf("failed to download file %s: %w", url, outcome.Err)
	}

	return outcome.Value.fullFileName, nil
}

// CleanLocalCache cleans the local cache folder.
func (m *MutexDownloadService) CleanLocalCache(fileAge time.Duration) error {
	m.resultListenersMutex.Lock()
	defer m.resultListenersMutex.Unlock()

	dirContent, err := os.ReadDir(m.cacheDirectory)
	if err != nil {
		return fmt.Errorf("failed to read cache dir content: %w", err)
	}

	currentlyDownloadingFiles := sliceutils.Map(
		maputils.Values(m.runningDownloads),
		func(value *RunningDownload, index int, slice []*RunningDownload) string {
			return value.TargetFileName
		},
	)

	now := time.Now()
	for _, entry := range dirContent {
		if sliceutils.Includes(currentlyDownloadingFiles, entry.Name()) {
			continue // Don't delete currently fetching files
		}

		info, err := entry.Info()
		if err != nil {
			return fmt.Errorf("failed to fetch file info for %s: %w", entry.Name(), err)
		}

		if info.ModTime().Add(fileAge).Before(now) { // If the mod time + the duration is still before now, delete it.
			if err := os.Remove(filepath.Join(m.cacheDirectory, entry.Name())); err != nil {
				return fmt.Errorf("failed to delete outdated cache %s: %w", entry.Name(), err)
			}
		}
	}

	return nil
}

func (m *MutexDownloadService) downloadURLToFile(ctx context.Context, url string, filename string) (string, error) {
	downloadTargetPath := path.Join(m.cacheDirectory, path.Base(filename))

	if err := DownloadURLTo(ctx, m.httpClient, url, downloadTargetPath); err != nil {
		return "", fmt.Errorf("failed to download: %w", err)
	}

	return downloadTargetPath, nil
}
