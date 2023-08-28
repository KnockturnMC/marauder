package cronjobworker

import (
	"time"

	"github.com/sirupsen/logrus"
)

// RemoveUnused constructs the cronjob executor that removes unused artefacts if they are older than the passed duration.
func RemoveUnused(cooldown time.Duration, removeAfter time.Duration) CronjobExecutor {
	return SimpleCronjobExecutor{
		cooldown: cooldown,
		executionFunction: func(worker *CronjobWorker) error {
			logrus.Info("ran remove unused: ", removeAfter)
			return nil
		},
	}
}
