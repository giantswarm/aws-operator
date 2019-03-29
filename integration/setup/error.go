// +build k8srequired

package setup

import (
	"strings"

	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/giantswarm/microerror"
)

var executionFailedError = &microerror.Error{
	Kind: "executionFailedError",
}

// IsExecutionFailed asserts executionFailedError.
func IsExecutionFailed(err error) bool {
	return microerror.Cause(err) == executionFailedError
}

var notExistsError = &microerror.Error{
	Kind: "notExistsError",
}

// IsNotExists asserts notExistsError.
func IsNotExists(err error) bool {
	c := microerror.Cause(err)

	if c == nil {
		return false
	}

	if strings.Contains(c.Error(), "does not exist") {
		return true
	}

	if c == notExistsError {
		return true
	}

	return false
}

var notFoundError = &microerror.Error{
	Kind: "notFoundError",
}

// IsNotFound asserts notFoundError.
func IsNotFound(err error) bool {
	return microerror.Cause(err) == notFoundError
}

var stackAlreadyExistsError = &microerror.Error{
	Kind: "stackAlreadyExistsError",
}

// IsStackAlreadyExists asserts alreadyExistsError.
func IsStackAlreadyExists(err error) bool {
	c := microerror.Cause(err)

	if c == nil {
		return false
	}

	if strings.Contains(c.Error(), cloudformation.ErrCodeAlreadyExistsException) {
		return true
	}

	if c == stackAlreadyExistsError {
		return true
	}

	return false
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
