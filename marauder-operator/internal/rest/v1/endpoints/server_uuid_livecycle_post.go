package endpoints

import (
	"fmt"
	"net/http"
	"strings"

	"gitea.knockturnmc.com/marauder/lib/pkg/controller"

	"gitea.knockturnmc.com/marauder/operator/pkg/manager"

	"gitea.knockturnmc.com/marauder/lib/pkg/models/networkmodel"
	"gitea.knockturnmc.com/marauder/lib/pkg/rest/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func ServerLifecycleActionPost(
	operatorIdentifier string,
	controllerClient controller.Client,
	serverManager manager.Manager,
) gin.HandlerFunc {
	return func(context *gin.Context) {
		serverUUIDAsString := context.Param("uuid")
		serverUUID, err := uuid.Parse(serverUUIDAsString)
		if err != nil {
			_ = context.Error(response.RestErrorFromDescription(http.StatusBadRequest, "could not parse uuid in url params"))
			return
		}

		action := networkmodel.LifecycleAction(strings.ToLower(context.Param("action")))
		if !networkmodel.KnownLifecycleChangeActionType(action) {
			_ = context.Error(response.RestErrorFromDescription(http.StatusBadRequest, fmt.Sprintf("unknown action %s", action)))
			return
		}

		server, err := controllerClient.FetchServer(context, serverUUID)
		if err != nil {
			_ = context.Error(response.RestErrorFromErr(
				http.StatusInternalServerError,
				fmt.Errorf("failed to fetch server %s: %w", serverUUIDAsString, err),
			))

			return
		}

		if server.OperatorRef.Identifier != operatorIdentifier {
			_ = context.Error(response.RestErrorFromDescription(
				http.StatusBadRequest,
				fmt.Sprintf("server %s is not managed by operator %s", serverUUID.String(), operatorIdentifier),
			))

			return
		}

		if handleLifecycleAction(context, action, serverManager, server) {
			context.Status(http.StatusOK)
		}
	}
}

// handleLifecycleAction handles the passed lifecycle action on the server.
func handleLifecycleAction(
	context *gin.Context,
	action networkmodel.LifecycleAction,
	serverManager manager.Manager,
	server networkmodel.ServerModel,
) bool {
	switch action {
	case networkmodel.Start:
		return handleLifecycleActionStart(context, serverManager, server)
	case networkmodel.Stop:
		return handleLifecycleActionStop(context, serverManager, server)
	case networkmodel.Restart:
		return handleLifecycleActionStop(context, serverManager, server) &&
			handleLifecycleActionStart(context, serverManager, server)
	case networkmodel.UpdateWithoutRestart, networkmodel.ForceUpdateWithoutRestart:
		return updateServerDeployments(context, serverManager, server, false, action == networkmodel.UpdateWithoutRestart)
	case networkmodel.UpdateWithRestart, networkmodel.ForceUpdateWithRestart:
		return handleLifecycleActionStop(context, serverManager, server) &&
			updateServerDeployments(context, serverManager, server, true, action == networkmodel.UpdateWithRestart) &&
			handleLifecycleActionStart(context, serverManager, server)
	default:
		_ = context.Error(response.RestErrorFromDescription(http.StatusInternalServerError, fmt.Sprintf("unhandled action %s", action)))
		return false
	}
}

// updateServerDeployments updates the deployments on a server via the passed server manager.
// If may only do this for artefacts that do not require a restart if requiresRestart is false.
func updateServerDeployments(
	context *gin.Context,
	serverManager manager.Manager,
	server networkmodel.ServerModel,
	restarting bool,
	failOnUnexpectedOldFilesOnDisk bool,
) bool {
	if err := serverManager.UpdateDeployments(context, server, restarting, failOnUnexpectedOldFilesOnDisk); err != nil {
		_ = context.Error(response.RestErrorFromKnownErr(map[error]response.KnownErr{
			manager.ErrServerRunning: {
				ResponseCode: http.StatusBadRequest, Description: fmt.Sprintf("the server %s is running", server.Name),
			},
		}, fmt.Errorf("failed to update deployments: %w", err)))

		return false
	}
	return true
}

// handleLifecycleActionStart handles the start lifecycle action.
func handleLifecycleActionStart(ctx *gin.Context, serverManager manager.Manager, server networkmodel.ServerModel) bool {
	if err := serverManager.Start(ctx, server); err != nil {
		_ = ctx.Error(response.RestErrorFromErr(http.StatusInternalServerError, fmt.Errorf("failed to start server: %w", err)))
		return false
	}

	return true
}

// handleLifecycleActionStart handles the stop lifecycle action.
func handleLifecycleActionStop(ctx *gin.Context, serverManager manager.Manager, server networkmodel.ServerModel) bool {
	if err := serverManager.Stop(ctx, server); err != nil {
		_ = ctx.Error(response.RestErrorFromErr(http.StatusInternalServerError, fmt.Errorf("failed to stop server: %w", err)))
		return false
	}

	return true
}
