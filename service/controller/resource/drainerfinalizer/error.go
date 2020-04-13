package drainerfinalizer

import (
	"strings"

	"github.com/giantswarm/microerror"
)

var invalidConfigError = &microerror.Error{
	Kind: "invalidConfigError",
}

// IsInvalidConfig asserts invalidConfigError.
func IsInvalidConfig(err error) bool {
	return microerror.Cause(err) == invalidConfigError
}

var missingAnnotationError = &microerror.Error{
	Kind: "missingAnnotationError",
}

func IsMissingAnnotationError(err error) bool {
	return microerror.Cause(err) == missingAnnotationError
}

var noActiveLifeCycleActionError = &microerror.Error{
	Kind: "noActiveLifeCycleActionError",
}

// IsNoActiveLifeCycleAction asserts noActiveLifeCycleActionError. It also
// checks for some string matching in the error message to figure if the AWS API
// gives the error we expect.
func IsNoActiveLifeCycleAction(err error) bool {
	c := microerror.Cause(err)

	if c == nil {
		return false
	}

	if strings.Contains(c.Error(), "no active life cycle action found") {
		return true
	}

	if c == noActiveLifeCycleActionError {
		return true
	}

	return false
}
