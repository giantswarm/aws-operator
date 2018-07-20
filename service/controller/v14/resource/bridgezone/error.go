package bridgezone

import "github.com/giantswarm/microerror"

var executionError = microerror.New("execution")

// IsExecution asserts executionError.
func IsExecution(err error) bool {
	return microerror.Cause(err) == executionError
}

var invalidConfigError = microerror.New("invalid config")

// IsInvalidConfig asserts invalidConfigError.
func IsInvalidConfig(err error) bool {
	return microerror.Cause(err) == invalidConfigError
}

var notFoundError = microerror.New("not found")

// IsNotFound asserts notFoundError.
func IsNotFound(err error) bool {
	return microerror.Cause(err) == notFoundError
}
