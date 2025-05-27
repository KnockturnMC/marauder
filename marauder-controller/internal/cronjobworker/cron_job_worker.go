package cronjobworker

import (
	"context"
	"fmt"
	"time"

	"github.com/knockturnmc/marauder/marauder-controller/internal/db/access"
	"github.com/knockturnmc/marauder/marauder-controller/pkg/cronjob"
	"github.com/knockturnmc/marauder/marauder-controller/sqlm"
	"github.com/knockturnmc/marauder/marauder-lib/pkg/operator"
	"github.com/sirupsen/logrus"
)

// CronjobExecutor defines a specific cronjobs executor that can be executed against a job worker parent.
type CronjobExecutor interface {
	// Execute executes the job executor.
	Execute(ctx context.Context, parent *CronjobWorker) error

	// Cooldown defines the cooldown of the executor
	Cooldown() time.Duration
}

type FetchedCronjob struct {
	Execution cronjob.Execution
	Executor  CronjobExecutor
}

// The CronjobWorker struct is the worker that is responsible for executing cronjobs for the controller.
type CronjobWorker struct {
	DB                  *sqlm.DB
	OperatorClientCache *operator.ClientCache
	cronjobs            map[cronjob.Type]*FetchedCronjob
	timerResetChan      chan time.Duration
}

// NewCronjobWorker constructs a new job worker for the given database and configuration.
func NewCronjobWorker(db *sqlm.DB, operatorClientCache *operator.ClientCache, executors map[cronjob.Type]CronjobExecutor) *CronjobWorker {
	preparedCronjobs := make(map[cronjob.Type]*FetchedCronjob)
	for cronjobType, executor := range executors {
		preparedCronjobs[cronjobType] = &FetchedCronjob{
			Execution: cronjob.Execution{
				NextExecution: time.Now().UTC(),
				Type:          cronjobType,
			},
			Executor: executor,
		}
	}

	return &CronjobWorker{
		DB:                  db,
		OperatorClientCache: operatorClientCache,
		cronjobs:            preparedCronjobs,
	}
}

// RescheduleCronjobAt schedules a new run of the worker in the given duration.
func (j *CronjobWorker) RescheduleCronjobAt(ctx context.Context, cronjobType cronjob.Type, duration time.Duration) error {
	if err := access.UpsertCronjobExecution(ctx, j.DB, cronjobType.Execution(time.Now().Add(duration).UTC())); err != nil {
		return fmt.Errorf("failed to upsert next cronjob execution: %w", err)
	}

	j.timerResetChan <- duration
	return nil
}

// Start starts the job worker.
func (j *CronjobWorker) Start(ctx context.Context) error {
	if len(j.cronjobs) == 0 {
		return nil
	}

	j.timerResetChan = make(chan time.Duration)
	timer := time.NewTimer(0 * time.Second)

	for {
		select {
		case <-ctx.Done():
			return nil
		case resetDuration := <-j.timerResetChan:
			timer.Reset(resetDuration)
		case <-timer.C:
			durationTillNext, err := j.runRunnableJobs(ctx)
			if err != nil {
				logrus.Error(fmt.Errorf("failed to run jobs: %w", err)) // sink error instead of breaking because of it.
				durationTillNext = 1 * time.Minute                      // Default to sensible back off
			}

			// Reset the timer, we run again when the next fetchedCronjob is available.
			timer.Reset(durationTillNext)
		}
	}
}

// runRunnableJobs attempts to run all runnable jobs.
func (j *CronjobWorker) runRunnableJobs(ctx context.Context) (time.Duration, error) {
	if err := j.updateFetchedCronjobsFromDB(ctx); err != nil {
		return 0, fmt.Errorf("failed to update local fetched cronjobs: %w", err)
	}

	timeNow := time.Now().UTC()
	for cronjobType, fetchedCronjob := range j.cronjobs {
		// Cronjob is not scheduled to be executed this run
		if fetchedCronjob.Execution.NextExecution.UTC().After(timeNow) {
			continue
		}

		// Execute the cronjob
		if err := fetchedCronjob.Executor.Execute(ctx, j); err != nil {
			return 0, fmt.Errorf("failed to execute cronjob %s: %w", cronjobType, err)
		}

		// Create a new execution and upsert it into the database.
		fetchedCronjob.Execution = cronjob.Execution{
			NextExecution: timeNow.Add(fetchedCronjob.Executor.Cooldown()),
			Type:          cronjobType,
		}
		if err := access.UpsertCronjobExecution(ctx, j.DB, fetchedCronjob.Execution); err != nil {
			return 0, fmt.Errorf("failed to update chron fetchedCronjob in db: %w", err)
		}

		logrus.Info("executed cronjob ", cronjobType, " at ", timeNow.UTC())
	}

	// Update the possible next execution
	earliestNextJobExecutionTime := timeNow
	for _, fetchedCronjob := range j.cronjobs {
		if time.Time.Equal(earliestNextJobExecutionTime, timeNow) || fetchedCronjob.Execution.NextExecution.UTC().Before(earliestNextJobExecutionTime) {
			earliestNextJobExecutionTime = fetchedCronjob.Execution.NextExecution.UTC()
		}
	}

	return earliestNextJobExecutionTime.Sub(timeNow), nil
}

// updateFetchedCronjobsFromDB updates the local fetched cronjobs to the database representation.
func (j *CronjobWorker) updateFetchedCronjobsFromDB(ctx context.Context) error {
	cronjobExecutions, err := access.FetchLastCronjobExecutions(ctx, j.DB)
	if err != nil {
		return fmt.Errorf("failed to fetch from sql: %w", err)
	}

	for _, execution := range cronjobExecutions {
		fetchedCronjob, ok := j.cronjobs[execution.Type]
		if !ok {
			continue
		}

		fetchedCronjob.Execution = execution
	}

	return nil
}
