package lifecycle

import (
	"strings"

	"github.com/giantswarm/microerror"
)

var executionFailedError = microerror.New("execution failed")

// IsExecutionFailed asserts executionFailedError.
func IsExecutionFailed(err error) bool {
	return microerror.Cause(err) == executionFailedError
}

var invalidConfigError = microerror.New("invalid config")

// IsInvalidConfig asserts invalidConfigError.
func IsInvalidConfig(err error) bool {
	return microerror.Cause(err) == invalidConfigError
}

var missingAnnotationError = microerror.New("missing annotation")

func IsMissingAnnotationError(err error) bool {
	return microerror.Cause(err) == missingAnnotationError
}

var noActiveLifecycleActionError = microerror.New("no active lifecycle action")

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
