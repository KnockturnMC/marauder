package cronjobworker

import (
	"context"
	"fmt"
	"time"

	"gitea.knockturnmc.com/marauder/controller/internal/db/access"
	"gitea.knockturnmc.com/marauder/controller/pkg/cronjob"
	"gitea.knockturnmc.com/marauder/controller/sqlm"
)

// CronjobExecutor defines a specific cronjobs executor that can be executed against a job worker parent.
type CronjobExecutor interface {
	// Execute executes the job executor.
	Execute(parent *CronjobWorker) error

	// Cooldown defines the cooldown of the executor
	Cooldown() time.Duration
}

type FetchedCronjob struct {
	Execution cronjob.Execution
	Executor  CronjobExecutor
}

// The CronjobWorker struct is the worker that is responsible for executing cronjobs for the controller.
type CronjobWorker struct {
	DB       *sqlm.DB
	cronjobs map[cronjob.Type]*FetchedCronjob
}

// Start starts the job worker.
func (j *CronjobWorker) Start(ctx context.Context) error {
	if len(j.cronjobs) == 0 {
		return nil
	}

	timer := time.NewTimer(0 * time.Second)
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-timer.C:
			durationTillNext, err := j.runRunnableJobs(ctx)
			if err != nil {
				return fmt.Errorf("failed to run jobs: %w", err)
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
	earliestNextJobExecutionTime := timeNow
	for cronjobType, fetchedCronjob := range j.cronjobs {
		if fetchedCronjob.Execution.NextExecution.UTC().Before(timeNow) {
			if err := fetchedCronjob.Executor.Execute(j); err != nil {
				return 0, fmt.Errorf("failed to execute cronjob %s: %w", cronjobType, err)
			}

			fetchedCronjob.Execution = cronjob.Execution{
				NextExecution: timeNow.Add(fetchedCronjob.Executor.Cooldown()),
				Type:          cronjobType,
			}
			if err := access.UpsertCronjobExecution(ctx, j.DB, fetchedCronjob.Execution); err != nil {
				return 0, fmt.Errorf("failed to update chron fetchedCronjob in db: %w", err)
			}
		}

		if earliestNextJobExecutionTime == timeNow || fetchedCronjob.Execution.NextExecution.Before(earliestNextJobExecutionTime) {
			earliestNextJobExecutionTime = fetchedCronjob.Execution.NextExecution
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

// NewCronjobWorker constructs a new job worker for the given database and configuration.
func NewCronjobWorker(db *sqlm.DB, executors map[cronjob.Type]CronjobExecutor) *CronjobWorker {
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

	return &CronjobWorker{DB: db, cronjobs: preparedCronjobs}
}
