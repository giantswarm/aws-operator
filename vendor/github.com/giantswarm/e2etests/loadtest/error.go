package loadtest

import "github.com/giantswarm/microerror"

var invalidConfigError = &microerror.Error{
	Kind: "invalidConfigError",
}

// IsInvalidConfig asserts invalidConfigError.
func IsInvalidConfig(err error) bool {
	return microerror.Cause(err) == invalidConfigError
}

var invalidExecutionError = &microerror.Error{
	Kind: "invalidExecutionError",
}

// IsInvalidExecution asserts invalidExecutionError.
func IsInvalidExecution(err error) bool {
	return microerror.Cause(err) == invalidExecutionError
}

var waitError = &microerror.Error{
	Kind: "waitError",
}

// IsWait asserts waitError.
func IsWait(err error) bool {
	return microerror.Cause(err) == waitError
}
