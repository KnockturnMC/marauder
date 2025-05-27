package endpoints

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/knockturnmc/marauder/marauder-lib/pkg/models/networkmodel"
)

// VersionGet generates the version endpoint for marauder controller.
func VersionGet(version string) gin.HandlerFunc {
	return func(context *gin.Context) {
		context.PureJSON(
			http.StatusOK,
			networkmodel.VersionResponseModel{Version: version},
		)
	}
}
