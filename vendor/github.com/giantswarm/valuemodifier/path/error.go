package path

import "github.com/giantswarm/microerror"

var invalidConfigError = &microerror.Error{
	Kind: "invalidConfigError",
}

// IsInvalidConfig asserts invalidConfigError.
func IsInvalidConfig(err error) bool {
	return microerror.Cause(err) == invalidConfigError
}

var invalidFormatError = &microerror.Error{
	Kind: "invalidFormatError",
}

// IsInvalidFormat asserts invalidFormatError.
func IsInvalidFormat(err error) bool {
	return microerror.Cause(err) == invalidFormatError
}

var keyNotIndexError = &microerror.Error{
	Kind: "keyNotIndexError",
}

// IsKeyNotIndex asserts keyNotIndexError.
func IsKeyNotIndex(err error) bool {
	return microerror.Cause(err) == keyNotIndexError
}

var notFoundError = &microerror.Error{
	Kind: "notFoundError",
}

// IsNotFound asserts notFoundError.
func IsNotFound(err error) bool {
	return microerror.Cause(err) == notFoundError
}
