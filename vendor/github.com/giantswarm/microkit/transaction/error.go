package transaction

import (
	"github.com/giantswarm/microerror"
)

var alreadyExistsError = microerror.New("already exists")

// IsAlreadyExists asserts alreadyExistsError.
func IsAlreadyExists(err error) bool {
	return microerror.Cause(err) == alreadyExistsError
}

var invalidConfigError = microerror.New("invalid config")

// IsInvalidConfig asserts invalidConfigError.
func IsInvalidConfig(err error) bool {
	return microerror.Cause(err) == invalidConfigError
}

var invalidExecutionError = microerror.New("invalid execution")

// IsInvalidExecution asserts invalidExecutionError.
func IsInvalidExecution(err error) bool {
	return microerror.Cause(err) == invalidExecutionError
}

var notFoundError = microerror.New("not found")

// IsNotFound asserts notFoundError.
func IsNotFound(err error) bool {
	return microerror.Cause(err) == notFoundError
}
