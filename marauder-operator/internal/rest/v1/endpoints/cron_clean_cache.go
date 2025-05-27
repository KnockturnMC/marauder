package endpoints

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/knockturnmc/marauder/marauder-lib/pkg/rest/response"
	"github.com/knockturnmc/marauder/marauder-lib/pkg/worker"
)

// CronCleanCache creates the endpoint that may be called to clear the operators cache.
func CronCleanCache(
	downloadService worker.DownloadService,
) gin.HandlerFunc {
	return func(context *gin.Context) {
		value := context.DefaultQuery("age", "1h")
		duration, err := time.ParseDuration(value)
		if err != nil {
			_ = context.Error(response.RestErrorFromErr(http.StatusBadRequest, fmt.Errorf("failed to parse age: %w", err)))
			return
		}

		if err := downloadService.CleanLocalCache(duration); err != nil {
			_ = context.Error(response.RestErrorFromErr(http.StatusInternalServerError, fmt.Errorf("failed to clear cache: %w", err)))
			return
		}
	}
}
