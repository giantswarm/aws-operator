package lifecycle

import "github.com/giantswarm/microerror"

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
