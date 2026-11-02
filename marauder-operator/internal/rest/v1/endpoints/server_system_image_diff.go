package endpoints

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/knockturnmc/marauder/marauder-lib/pkg/rest/response"
	"github.com/knockturnmc/marauder/marauder-operator/pkg/manager"
)

// ServerSystemImageDetails yields the version of a local image in the container registry.
func ServerSystemImageDetails(
	serverManager manager.Manager,
) gin.HandlerFunc {
	return func(context *gin.Context) {
		server, err := getServerModelKey(context)
		if err != nil {
			_ = context.Error(response.RestErrorFromErr(http.StatusInternalServerError, err))
			return
		}

		imageDetails, err := serverManager.FetchServerImageDetails(context, server)
		if err != nil {
			_ = context.Error(response.RestErrorFromErr(
				http.StatusInternalServerError,
				fmt.Errorf("failed to compare server image version %s: %w", server.UUID, err),
			))

			return
		}

		context.JSONP(http.StatusOK, imageDetails)
	}
}
