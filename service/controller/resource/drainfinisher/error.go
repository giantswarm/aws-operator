package drainfinisher

import (
	"strings"

	"github.com/giantswarm/microerror"
)

// executionFailedError is an error type for situations where Resource execution
// cannot continue and must always fall back to operatorkit.
//
// This error should never be matched against and therefore there is no matcher
// implement. For further information see:
//
//     https://github.com/giantswarm/fmt/blob/master/go/errors.md#matching-errors
//
var executionFailedError = &microerror.Error{
	Kind: "executionFailedError",
}

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

var noActiveLifecycleActionError = &microerror.Error{
	Kind: "noActiveLifecycleActionError",
}

// IsNoActiveLifecycleAction asserts noActiveLifecycleActionError. It also
// checks for some string matching in the error message to figure if the AWS API
// gives the error we expect.
func IsNoActiveLifecycleAction(err error) bool {
	c := microerror.Cause(err)

	if c == nil {
		return false
	}

	if strings.Contains(c.Error(), "No active Lifecycle Action found") {
		return true
	}

	if c == noActiveLifecycleActionError {
		return true
	}

	return false
}
