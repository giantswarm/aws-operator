package setup

import (
	"strings"

	"github.com/giantswarm/microerror"
)

var notFoundError = &microerror.Error{
	Kind: "notFoundError",
}

// IsNotFound asserts notFoundError.
func IsNotFound(err error) bool {
	return microerror.Cause(err) == notFoundError
}

var stackNotFoundError = &microerror.Error{
	Kind: "stackNotFoundError",
}

// IsStackNotFound asserts stackNotFoundError.
// Copied from service/controller/v17/cloudformation/error.go.
func IsStackNotFound(err error) bool {
	if err == nil {
		return false
	}

	if strings.Contains(microerror.Cause(err).Error(), "does not exist") {
		return true
	}

	if microerror.Cause(err) == stackNotFoundError {
		return true
	}

	return false

}

var stillExistsError = &microerror.Error{
	Kind: "stillExistsError",
}

// IsStillExists asserts stillExistsError.
func IsStillExists(err error) bool {
	return microerror.Cause(err) == stillExistsError
}

var tooManyResultsError = &microerror.Error{
	Kind: "tooManyResultsError",
}

// IsTooManyResults asserts tooManyResultsError.
func IsTooManyResults(err error) bool {
	return microerror.Cause(err) == tooManyResultsError
}
