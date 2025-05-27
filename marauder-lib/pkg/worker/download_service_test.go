package worker_test

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/knockturnmc/marauder/marauder-lib/pkg/worker"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/phayes/freeport"
)

func downloadServiceDispatcher() *worker.Dispatcher[worker.DownloadResult] {
	dispatcher, err := worker.NewDispatcher[worker.DownloadResult](5)
	Expect(err).To(Not(HaveOccurred()))

	return dispatcher
}

var _ = Describe("DownloadService", Label("functiontest"), func() {
	var (
		port          int
		httpServerMux *http.ServeMux
		httpServer    *http.Server
		tempDirFolder string
	)

	_ = BeforeEach(func() {
		var err error
		tempDirFolder, err = os.MkdirTemp("", fmt.Sprintf("download-service-test-%d", GinkgoParallelProcess()))
		Expect(err).To(Not(HaveOccurred()))

		freePort, err := freeport.GetFreePort()
		Expect(err).To(Not(HaveOccurred()))

		port = freePort
		httpServerMux = http.NewServeMux()

		httpServer = &http.Server{
			Addr:              net.JoinHostPort("localhost", strconv.Itoa(port)),
			Handler:           httpServerMux,
			ReadHeaderTimeout: time.Minute,
		}
		go func() {
			_ = httpServer.ListenAndServe()
		}()
	})

	_ = AfterEach(func() {
		Expect(os.RemoveAll(tempDirFolder)).To(Not(HaveOccurred()))
		Expect(httpServer.Close()).To(Not(HaveOccurred()))
	})

	_ = Describe("downloading a file", func() {
		Context("from a responsive web server", func() {
			It("should download the file", func() {
				httpServerMux.HandleFunc("/good", func(writer http.ResponseWriter, request *http.Request) {
					writer.WriteHeader(http.StatusOK)
					_, err := writer.Write([]byte("good"))
					Expect(err).To(Not(HaveOccurred()))
				})

				downloadService := worker.NewMutexDownloadService(http.DefaultClient, downloadServiceDispatcher(), tempDirFolder)
				download, err := downloadService.Download(context.Background(), fmt.Sprintf("http://%s/good", httpServer.Addr), "good.txt")

				Expect(err).To(Not(HaveOccurred()))
				downloadedContent, err := os.ReadFile(filepath.Clean(download))
				Expect(err).To(Not(HaveOccurred()))
				Expect(string(downloadedContent)).To(BeEquivalentTo("good"))
			})

			It("should be able to reuse download calls currently performed", func() {
				counter := atomic.Int64{}
				httpServerMux.HandleFunc("/good", func(writer http.ResponseWriter, request *http.Request) {
					time.Sleep(50 * time.Millisecond)

					counter.Add(1)
					writer.WriteHeader(http.StatusOK)
					_, err := writer.Write([]byte("good"))
					Expect(err).To(Not(HaveOccurred()))
				})

				resultChan := make(chan bool)
				parallelCount := 3
				downloadService := worker.NewMutexDownloadService(http.DefaultClient, downloadServiceDispatcher(), tempDirFolder)

				for range parallelCount {
					go func() {
						download, err := downloadService.Download(context.Background(), fmt.Sprintf("http://%s/good", httpServer.Addr), "good.txt")
						Expect(err).To(Not(HaveOccurred()))

						Expect(err).To(Not(HaveOccurred()))
						downloadedContent, err := os.ReadFile(filepath.Clean(download))
						Expect(err).To(Not(HaveOccurred()))
						Expect(string(downloadedContent)).To(BeEquivalentTo("good"))

						resultChan <- true
					}()
				}

				for range parallelCount {
					Eventually(resultChan).Should(Receive())
				}

				Expect(counter.Load()).To(BeEquivalentTo(1)) // should all be cached
			})
		})

		Context("from an unresponsive web server", func() {
			It("should fail with the passed context", func() {
				httpServerMux.HandleFunc("/timeout", func(writer http.ResponseWriter, request *http.Request) {
					time.Sleep(50 * time.Millisecond)
				})

				downloadService := worker.NewMutexDownloadService(http.DefaultClient, downloadServiceDispatcher(), tempDirFolder)

				timeoutContext, cancelFunc := context.WithTimeout(context.Background(), 10*time.Millisecond)
				defer cancelFunc()

				_, err := downloadService.Download(timeoutContext, fmt.Sprintf("http://%s/timeout", httpServer.Addr), "timeout.txt")
				Expect(err).To(MatchError(context.DeadlineExceeded))
			})
		})
	})
})
