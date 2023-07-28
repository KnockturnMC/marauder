package utils

import "errors"

// CheckDockerError checks if the given error or any of its unwrapped inners match the given function.
// This method is needed as the docker client does not yet make use of the new error logic present in go.
func CheckDockerError(err error, matcher func(err error) bool) bool {
	for err != nil {
		if matcher(err) {
			return true
		}

		err = errors.Unwrap(err)
	}

	return false
}
