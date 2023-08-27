package endpoints

import (
	"database/sql"
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"

	"gitea.knockturnmc.com/marauder/controller/internal/db/access"
	"gitea.knockturnmc.com/marauder/controller/sqlm"
	"gitea.knockturnmc.com/marauder/lib/pkg/rest/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func OperationServerProxy(
	db *sqlm.DB,
	operatorClient *http.Client,
	protocol string,
) gin.HandlerFunc {
	return func(context *gin.Context) {
		serverUUIDAsString := context.Param("server")
		serverUUID, err := uuid.Parse(serverUUIDAsString)
		if err != nil {
			_ = context.Error(response.RestErrorFromDescription(http.StatusBadRequest, "could not serverUUID server uuid "+err.Error()))
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

		// Construct request to operator endpoint
		controllerEndpoint := fmt.Sprintf(
			"%s://%s/v1%s",
			protocol,
			net.JoinHostPort(server.OperatorRef.Host, strconv.Itoa(server.OperatorRef.Port)),
			context.Param("path"),
		)
		request, err := http.NewRequestWithContext(
			context,
			context.Request.Method,
			controllerEndpoint,
			context.Request.Body,
		)
		if err != nil {
			_ = context.Error(response.RestErrorFromErr(
				http.StatusInternalServerError,
				fmt.Errorf("failed to create http request to controller: %w", err),
			))

			return
		}

		// Execute request
		operatorResp, err := operatorClient.Do(request)
		if err != nil {
			_ = context.Error(response.RestErrorFromErr(
				http.StatusInternalServerError,
				fmt.Errorf("failed to perform request to operator: %w", err),
			))

			return
		}

		defer func() { _ = operatorResp.Body.Close() }()

		// Copy headers
		for headerKey, headerVal := range operatorResp.Header {
			for _, headerValInSlice := range headerVal {
				context.Writer.Header().Add(headerKey, headerValInSlice)
			}
		}

		context.Status(operatorResp.StatusCode)
		if _, err := io.Copy(context.Writer, operatorResp.Body); err != nil {
			_ = context.Error(response.RestErrorFromErr(http.StatusInternalServerError, fmt.Errorf("failed to copy operator response: %w", err)))
			return
		}
	}
}
