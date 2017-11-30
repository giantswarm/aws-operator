package service

import (
	"github.com/giantswarm/microerror"
)

var (
	invalidConfigError  = microerror.New("invalid config")
	invalidVersionError = microerror.New("invalid version")
)

// IsInvalidConfig asserts invalidConfigError.
func IsInvalidConfig(err error) bool {
	return microerror.Cause(err) == invalidConfigError
}

// IsInvalidVersion asserts invalidVersionError.
func IsInvalidVersion(err error) bool {
	return microerror.Cause(err) == invalidVersionError
}
