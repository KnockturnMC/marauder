package endpoints

import (
	"net/http"

	"gitea.knockturnmc.com/marauder/lib/pkg/networkmodel"

	"github.com/gin-gonic/gin"
)

// VersionEndpoint generates the version endpoint for marauder controller.
func VersionEndpoint(version string) gin.HandlerFunc {
	return func(context *gin.Context) {
		context.PureJSON(
			http.StatusOK,
			networkmodel.VersionResponseModel{Version: version},
		)
	}
}
