package response

import (
	"errors"

	"github.com/google/uuid"
)

// ErrDescriptiveRequestError is the base error for any purely descriptive error response.
var ErrDescriptiveRequestError = errors.New("server rest error")

// The RestRequestError contains a middleware that occurred while processing a request.
type RestRequestError struct {
	Description   string `json:"description,omitempty"`
	Identifier    string `json:"identifier"`
	internalError error

	responseCode int
}

// Error implements the error type by converting the error to a string.
func (e RestRequestError) Error() string {
	return e.internalError.Error()
}

// Unwrap unwraps the error request into its inner error.
func (e RestRequestError) Unwrap() error {
	return e.internalError
}

// ResponseCode returns the response code the error indicates.
func (e RestRequestError) ResponseCode() int {
	return e.responseCode
}

// RestErrorFromErr creates a new middleware response from a passed go middleware.
func RestErrorFromErr(responseCode int, err error) *RestRequestError {
	return RestErrorFrom(responseCode, "", err)
}

// RestErrorFrom creates a new middleware response from a passed go middleware and the passed description.
func RestErrorFrom(responseCode int, description string, err error) *RestRequestError {
	return &RestRequestError{
		Description:   description,
		Identifier:    uuid.New().String(),
		internalError: err,
		responseCode:  responseCode,
	}
}

// RestErrorFromDescription creates a new middleware response from the passed description.
func RestErrorFromDescription(responseCode int, description string) *RestRequestError {
	return RestErrorFrom(responseCode, description, ErrDescriptiveRequestError)
}
