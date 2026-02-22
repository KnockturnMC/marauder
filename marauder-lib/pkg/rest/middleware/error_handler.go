package middleware

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/knockturnmc/marauder/marauder-lib/pkg/rest/response"
	"github.com/sirupsen/logrus"
)

// ErrorHandler creates the middleware handler middleware.
func ErrorHandler() gin.HandlerFunc {
	return func(context *gin.Context) {
		context.Next() // Error handling is post request stuff

		postHandleErr := make([]*gin.Error, 0, len(context.Errors))

		for _, requestErr := range context.Errors {
			var aviorErr *response.RestRequestError
			if !errors.As(requestErr.Err, &aviorErr) {
				postHandleErr = append(postHandleErr, requestErr)
				continue
			}

			logger := logrus.WithFields(map[string]any{
				"errorIdentifier": aviorErr.Identifier,
				"endpointUri":     context.Request.RequestURI,
				"clientIP":        context.ClientIP(),
			})
			logger.Logf(logLevelForRequestErr(aviorErr), "err: %s, desc: %s", aviorErr.Error(), aviorErr.Description)

			context.PureJSON(aviorErr.ResponseCode(), aviorErr)
		}

		context.Errors = postHandleErr
	}
}

// logLevelForRequestErr attempts to find the proper log level for a specific request error.
func logLevelForRequestErr(requestErr *response.RestRequestError) logrus.Level {
	if errors.Is(requestErr.Unwrap(), response.ErrDescriptiveRequestError) {
		return logrus.WarnLevel
	}

	if requestErr.ResponseCode() < 500 {
		return logrus.WarnLevel
	}

	return logrus.ErrorLevel
}
