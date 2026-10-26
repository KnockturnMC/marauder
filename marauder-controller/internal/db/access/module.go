package access

import (
	"errors"
	"net/http"
	"strings"

	"github.com/lib/pq"
)

// RestErrFromAccessErr attempts to guess a http status code for a pq based error based on the reported postgres error.
func RestErrFromAccessErr(err error) int {
	if potentialErr, ok := errors.AsType[*pq.Error](err); ok {
		if strings.HasPrefix(potentialErr.Constraint, "un") {
			return http.StatusConflict
		}
	}

	return http.StatusInternalServerError
}
