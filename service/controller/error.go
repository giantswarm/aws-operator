package controller

import (
	"github.com/giantswarm/microerror"
)

var invalidConfigError = &microerror.Error{
	Kind: "invalidConfigError",
}

// IsInvalidConfig asserts invalidConfigError.
func IsInvalidConfig(err error) bool {
	return microerror.Cause(err) == invalidConfigError
}

var invalidVersionError = &microerror.Error{
	Kind: "invalidVersionError",
}

// IsInvalidVersion asserts invalidVersionError.
func IsInvalidVersion(err error) bool {
	return microerror.Cause(err) == invalidVersionError
}
