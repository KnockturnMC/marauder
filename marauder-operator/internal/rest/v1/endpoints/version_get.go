package endpoints

import (
	"net/http"

	"gitea.knockturnmc.com/marauder/lib/pkg/models/networkmodel"

	"github.com/gin-gonic/gin"
)

// VersionGet generates the version endpoint for marauder operator.
func VersionGet(version string) gin.HandlerFunc {
	return func(context *gin.Context) {
		context.PureJSON(
			http.StatusOK,
			networkmodel.VersionResponseModel{Version: version},
		)
	}
}
