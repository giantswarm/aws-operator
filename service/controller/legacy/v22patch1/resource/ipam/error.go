package ipam

import "github.com/giantswarm/microerror"

var invalidConfigError = &microerror.Error{
	Kind: "invalid config",
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
