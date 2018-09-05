package bridgezone

import "github.com/giantswarm/microerror"

var executionError = &microerror.Error{
	Kind: "executionError",
}

// IsExecution asserts executionError.
func IsExecution(err error) bool {
	return microerror.Cause(err) == executionError
}

var invalidConfigError = &microerror.Error{
	Kind: "invalidConfigError",
}

// IsInvalidConfig asserts invalidConfigError.
func IsInvalidConfig(err error) bool {
	return microerror.Cause(err) == invalidConfigError
}

var notFoundError = &microerror.Error{
	Kind: "notFoundError",
}

// IsNotFound asserts notFoundError.
func IsNotFound(err error) bool {
	return microerror.Cause(err) == notFoundError
}
