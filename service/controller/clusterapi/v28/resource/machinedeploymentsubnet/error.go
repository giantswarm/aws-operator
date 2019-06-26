package machinedeploymentsubnet

import "github.com/giantswarm/microerror"

var executionFailedError = &microerror.Error{
	Kind: "executionFailedError",
}

// IsExecutionFailed asserts executionFailedError.
func IsExecutionFailed(err error) bool {
	return microerror.Cause(err) == executionFailedError
}

var invalidConfigError = &microerror.Error{
	Kind: "invalid config",
}

// IsInvalidConfig asserts invalidConfigError.
func IsInvalidConfig(err error) bool {
	return microerror.Cause(err) == invalidConfigError
}
