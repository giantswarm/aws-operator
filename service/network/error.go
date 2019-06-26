package network

import "github.com/giantswarm/microerror"

var invalidConfigError = &microerror.Error{
	Kind: "invalidConfigError",
}

// IsInvalidConfig asserts invalidConfigError.
func IsInvalidConfig(err error) bool {
	return microerror.Cause(err) == invalidConfigError
}

var invalidParameterError = &microerror.Error{
	Kind: "invalid parameter",
}

// IsInvalidParameter asserts invalidParameterError.
func IsInvalidParameter(err error) bool {
	return microerror.Cause(err) == invalidParameterError
}
