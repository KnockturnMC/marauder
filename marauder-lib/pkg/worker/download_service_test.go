package worker_test

import (
	"net"
	"net/http"
	"strconv"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/phayes/freeport"
)

var _ = Describe("DownloadService", Label("unittest"), func() {
	var (
		port          int
		httpServerMux *http.ServeMux
		httpServer    *http.Server
	)

	_ = BeforeEach(func() {
		freePort, err := freeport.GetFreePort()
		Expect(err).To(Not(HaveOccurred()))

		port = freePort
		httpServerMux = http.NewServeMux()
		httpServer := &http.Server{
			Addr:              net.JoinHostPort("localhost", strconv.Itoa(port)),
			Handler:           httpServerMux,
			ReadHeaderTimeout: time.Minute,
		}
		go func() {
			Expect(httpServer.ListenAndServe()).To(Not(HaveOccurred()))
		}()
	})

	_ = AfterEach(func() {
		Expect(httpServer.Close()).To(Not(HaveOccurred()))
	})
})
