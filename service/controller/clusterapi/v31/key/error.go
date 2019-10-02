package key

import "github.com/giantswarm/microerror"

var invalidParameterError = &microerror.Error{
	Kind: "invalid parameter",
}

// IsInvalidParameter asserts invalidParameterError.
func IsInvalidParameter(err error) bool {
	return microerror.Cause(err) == invalidParameterError
}

var notFoundError = &microerror.Error{
	Kind: "notFoundError",
}

// IsNotFound asserts notFoundError.
func IsNotFound(err error) bool {
	return microerror.Cause(err) == notFoundError
}

var wrongTypeError = &microerror.Error{
	Kind: "wrongTypeError",
}

// IsWrongTypeError asserts wrongTypeError.
func IsWrongTypeError(err error) bool {
	return microerror.Cause(err) == wrongTypeError
}
