package daemon

import (
	"github.com/giantswarm/microerror"
)

var invalidConfigError = microerror.New("invalid config")

// IsInvalidConfig asserts invalidConfigError.
func IsInvalidConfig(err error) bool {
	return microerror.Cause(err) == invalidConfigError
}

var invalidFlagError = microerror.New("invalid flag")

// IsInvalidFlag asserts invalidFlagError.
func IsInvalidFlag(err error) bool {
	return microerror.Cause(err) == invalidFlagError
}
