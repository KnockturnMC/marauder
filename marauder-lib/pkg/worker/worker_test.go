package worker_test

import (
	"time"

	"gitea.knockturnmc.com/marauder/lib/pkg/worker"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Worker", Label("unittest"), func() {
	_ = Describe("Creating a dispatcher", func() {
		Context("with 0 or less workers", func() {
			It("should fail to create the dispatcher", func() {
				_, err := worker.NewDispatcher[string](0)
				Expect(err).To(MatchError(worker.ErrNotEnoughWorkers))
			})
		})
	})

	_ = Describe("Scheduling work on a worker", func() {
		Context("with a non-erroring work function", func() {
			It("should properly finish the work and publish it to the channel", func() {
				dispatcher, err := worker.NewDispatcher[string](1)
				Expect(err).To(Not(HaveOccurred()))

				dispatch := dispatcher.Dispatch(func() (string, error) {
					time.Sleep(20 * time.Millisecond)
					return "this worked!", nil
				})

				var outcome worker.Outcome[string]
				Eventually(dispatch).Should(Receive(&outcome))
				Expect(outcome.Err).To(Not(HaveOccurred()))
				Expect(outcome.Value).To(BeEquivalentTo("this worked!"))
			})

			It("should properly accept multiple (10) work loads", func() {
				workCount := 10
				workerCount := 2
				sleep := 5 * time.Millisecond

				dispatcher, err := worker.NewDispatcher[string](workerCount)
				Expect(err).To(Not(HaveOccurred()))

				work := func() (string, error) {
					time.Sleep(sleep)
					return "this worked!", nil
				}

				begin := time.Now()
				workSlice := make([]<-chan worker.Outcome[string], 0)
				for range workerCount {
					workSlice = append(workSlice, dispatcher.Dispatch(work))
				}
				for _, workChan := range workSlice {
					var outcome worker.Outcome[string]
					Eventually(workChan).Should(Receive(&outcome))
					Expect(outcome.Err).To(Not(HaveOccurred()))
					Expect(outcome.Value).To(BeEquivalentTo("this worked!"))
				}

				Expect(time.Now()).To(BeTemporally("~", begin.Add(sleep*time.Duration(workCount/workerCount)), 20*time.Millisecond))
			})
		})
	})
})
