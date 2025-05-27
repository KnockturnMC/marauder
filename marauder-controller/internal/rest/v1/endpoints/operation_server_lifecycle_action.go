package endpoints

import (
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/knockturnmc/marauder/marauder-controller/internal/cronjobworker"
	"github.com/knockturnmc/marauder/marauder-controller/internal/db/access"
	"github.com/knockturnmc/marauder/marauder-controller/pkg/cronjob"
	"github.com/knockturnmc/marauder/marauder-controller/sqlm"
	"github.com/knockturnmc/marauder/marauder-lib/pkg/models/networkmodel"
	"github.com/knockturnmc/marauder/marauder-lib/pkg/operator"
	"github.com/knockturnmc/marauder/marauder-lib/pkg/rest/response"
)

func OperationServerLifecycleAction(
	db *sqlm.DB,
	operatorClientCache *operator.ClientCache,
	cronjobWorkerRef *cronjobworker.CronjobWorker,
) gin.HandlerFunc {
	return func(context *gin.Context) {
		serverUUIDAsString := context.Param("server")
		serverUUID, err := uuid.Parse(serverUUIDAsString)
		if err != nil {
			_ = context.Error(response.RestErrorFromDescription(http.StatusBadRequest, "could not serverUUID server uuid "+err.Error()))
			return
		}

		// parse lifecycle action
		lifecycleAction := networkmodel.LifecycleAction(context.Param("action"))
		if !networkmodel.KnownLifecycleChangeActionType(lifecycleAction) {
			_ = context.Error(response.RestErrorFromDescription(http.StatusBadRequest, fmt.Sprintf("unknown lifecycle action %s", lifecycleAction)))
			return
		}

		// Fetch server
		server, err := access.FetchServer(context, db, serverUUID)
		if err != nil {
			_ = context.Error(response.RestErrorFromKnownErr(
				map[error]response.KnownErr{
					sql.ErrNoRows: {ResponseCode: http.StatusNotFound, Description: "could not find server " + serverUUID.String()},
				},
				fmt.Errorf("failed to fetch server %s: %w", serverUUID, err)),
			)

			return
		}

		delayedByAsString, delayedByFound := context.GetQuery("delay")
		delayedByParsed, err := time.ParseDuration(delayedByAsString)
		if err != nil && delayedByFound {
			_ = context.Error(response.RestErrorFromErr(http.StatusBadRequest, fmt.Errorf("failed to parse delay %s: %w", delayedByAsString, err)))

			return
		}

		scheduledLifecycleAction := networkmodel.ScheduledLifecycleAction{
			UUID:            uuid.New(),
			ServerUUID:      server.UUID,
			Server:          server,
			LifecycleAction: lifecycleAction,
			TimeOfExecution: time.Now().UTC().Add(delayedByParsed),
		}

		// if no delay was specified, we execute the lifecycle action directly.
		if !delayedByFound {
			operatorClient := operatorClientCache.GetOrCreateFromRef(server.OperatorRef)
			if err := operatorClient.ExecuteLifecycleAction(context, serverUUID, lifecycleAction); err != nil {
				_ = context.Error(response.RestErrorFromErr(
					http.StatusInternalServerError,
					fmt.Errorf("failed to execute lifecycle action %s: %w", lifecycleAction, err),
				))

				return
			}

			context.JSONP(http.StatusOK, scheduledLifecycleAction)
			return
		}

		scheduledLifecycleAction, err = access.InsertOrMergeScheduledLifecycleAction(
			context,
			db,
			scheduledLifecycleAction,
		)
		if err != nil {
			_ = context.Error(response.RestErrorFromErr(
				access.RestErrFromAccessErr(err),
				fmt.Errorf("failed to insert scheduled lifecycle action: %w", err),
			))
			return
		}

		// Schedule the execution cronjob down the line to run the scheduled action inserted above.
		if err := cronjobWorkerRef.RescheduleCronjobAt(
			context,
			cronjob.ExecuteScheduledLifecycleActionsIdentifier,
			scheduledLifecycleAction.TimeOfExecution.Sub(time.Now().UTC()),
		); err != nil {
			_ = context.Error(response.RestErrorFromErr(http.StatusInternalServerError, fmt.Errorf("failed to reschedule cronjob: %w", err)))
			return
		}

		context.JSONP(http.StatusOK, scheduledLifecycleAction)
	}
}
