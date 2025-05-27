package endpoints

import (
	"database/sql"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/knockturnmc/marauder/marauder-controller/internal/db/access"
	"github.com/knockturnmc/marauder/marauder-controller/sqlm"
	"github.com/knockturnmc/marauder/marauder-lib/pkg/operator"
	"github.com/knockturnmc/marauder/marauder-lib/pkg/rest/response"
)

func OperationServerProxy(
	db *sqlm.DB,
	operatorClientCache *operator.ClientCache,
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

		// Fetch operator client
		operatorClient := operatorClientCache.GetOrCreate(server.OperatorIdentifier, server.OperatorRef.Host, server.OperatorRef.Port)

		// Execute request
		operatorResp, err := operatorClient.DoHTTPRequest(
			context,
			context.Request.Method,
			context.Param("path"),
			context.Request.Body,
			operator.None,
		)
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
