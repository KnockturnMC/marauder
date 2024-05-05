package worker

import (
	"errors"
	"fmt"
)

// ErrNotEnoughWorkers is returned if a dispatcher is attempted to be generated without worker.
var ErrNotEnoughWorkers = errors.New("cannot create dispatcher without workers")

// Outcome represents the finished outcome of a dispatchers work queue.
type Outcome[R any] struct {
	Err   error
	Value R
}

// Work represents work to be done on a worker.
type Work[T any] func() (T, error)

// Dispatcher is responsible for dispatching new work to its worker pool.
type Dispatcher[T any] struct {
	availableWorkers chan Worker[T]
}

// NewDispatcher constructs a new dispatcher with the given worker count.
func NewDispatcher[T any](workerCount int) (*Dispatcher[T], error) {
	if workerCount < 1 {
		return nil, fmt.Errorf("cannot generate dispatcher with worker count: %d: %w", workerCount, ErrNotEnoughWorkers)
	}

	availableWorkerChan := make(chan Worker[T], workerCount)

	for range workerCount {
		availableWorkerChan <- Worker[T]{}
	}

	return &Dispatcher[T]{
		availableWorkers: availableWorkerChan,
	}, nil
}

// Dispatch dispatches the work to the dispatcher, returning a channel that concludes to an outcome once the worker finishes.
func (d *Dispatcher[T]) Dispatch(work Work[T]) <-chan Outcome[T] {
	resultChannel := make(chan Outcome[T], 1) // buffer the result, the worker can return
	go func() {
		worker := <-d.availableWorkers // pull available worker for worker queue
		worker.DoWork(work, resultChannel)
		d.availableWorkers <- worker // resubmit the worker to the worker queue. Another dispatch may then grab it
	}()

	return resultChannel
}

type Worker[T any] struct{}

// DoWork executes the work of the worker.
func (w *Worker[T]) DoWork(work Work[T], callback chan<- Outcome[T]) {
	outcome, err := work()
	callback <- Outcome[T]{
		Err:   err,
		Value: outcome,
	}
	close(callback)
}
