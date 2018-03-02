package cloudformation

import "github.com/giantswarm/microerror"

var invalidConfigError = microerror.New("invalid config")

// IsInvalidConfig asserts invalidConfigError.
func IsInvalidConfig(err error) bool {
	return microerror.Cause(err) == invalidConfigError
}

var outputNotFoundError = microerror.New("output not found")

// IsOutputNotFound asserts outputNotFoundError.
func IsOutputNotFound(err error) bool {
	return microerror.Cause(err) == outputNotFoundError
}
