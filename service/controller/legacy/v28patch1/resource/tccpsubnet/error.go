package tccpsubnet

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

var vpcNotFoundError = &microerror.Error{
	Kind: "vpcNotFoundError",
}

// IsVPCNotFound asserts vpcNotFoundError.
func IsVPCNotFound(err error) bool {
	return microerror.Cause(err) == vpcNotFoundError
}
