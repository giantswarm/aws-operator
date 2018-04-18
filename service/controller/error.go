package controller

import (
	"github.com/giantswarm/microerror"
)

var invalidConfigError = microerror.New("invalid config")

// IsInvalidConfig asserts invalidConfigError.
func IsInvalidConfig(err error) bool {
	return microerror.Cause(err) == invalidConfigError
}

var invalidVersionError = microerror.New("invalid version")

// IsInvalidVersion asserts invalidVersionError.
func IsInvalidVersion(err error) bool {
	return microerror.Cause(err) == invalidVersionError
}
