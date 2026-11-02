package endpoints

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/knockturnmc/marauder/marauder-lib/pkg/controller"
	"github.com/knockturnmc/marauder/marauder-lib/pkg/models/networkmodel"
	"github.com/knockturnmc/marauder/marauder-lib/pkg/rest/response"
)

const ServerModelKey = "server-model-key"

var ErrContextKeyNotFound = errors.New("could not find key in context")

// getServerModelKey yields the server model from the context or an error if none was found.
func getServerModelKey(ctx *gin.Context) (networkmodel.ServerModel, error) {
	server, found := ctx.Get(ServerModelKey)
	if !found {
		return networkmodel.ServerModel{}, ErrContextKeyNotFound
	}

	cast, castable := server.(networkmodel.ServerModel)
	if !castable {
		return networkmodel.ServerModel{}, ErrContextKeyNotFound
	}

	return cast, nil
}

func ServerUUIDMiddleware(
	operatorIdentifier string,
	controllerClient controller.Client,
) func(*gin.Context) {
	return func(context *gin.Context) {
		serverUUIDAsString := context.Param("uuid")
		serverUUID, err := uuid.Parse(serverUUIDAsString)
		if err != nil {
			_ = context.Error(response.RestErrorFromDescription(http.StatusBadRequest, "could not parse uuid in url params"))
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

		context.Set(ServerModelKey, server)
		context.Next()
	}
}
