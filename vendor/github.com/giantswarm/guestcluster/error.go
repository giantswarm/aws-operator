package guestcluster

import (
	"github.com/giantswarm/microerror"
)

var invalidConfigError = microerror.New("invalid config")

// IsInvalidConfig asserts invalidConfigError.
func IsInvalidConfig(err error) bool {
	return microerror.Cause(err) == invalidConfigError
}

var timeoutError = microerror.New("timeout")

// IsTimeout asserts timeoutError.
func IsTimeout(err error) bool {
	return microerror.Cause(err) == timeoutError
}
