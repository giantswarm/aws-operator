package apprclient

import "github.com/giantswarm/microerror"

var invalidConfigError = &microerror.Error{
	Kind: "invalidConfigError",
}

// IsInvalidConfig asserts invalidConfigError.
func IsInvalidConfig(err error) bool {
	return microerror.Cause(err) == invalidConfigError
}

var invalidStatusCodeError = &microerror.Error{
	Kind: "invalidStatusCodeError",
}

// IsInvalidStatusCode asserts invalidStatusCodeError.
func IsInvalidStatusCode(err error) bool {
	return microerror.Cause(err) == invalidStatusCodeError
}

var unknownStatusError = &microerror.Error{
	Kind: "unknownStatusError",
}

// IsUnknownStatus asserts unknownStatusError.
func IsUnknownStatus(err error) bool {
	return microerror.Cause(err) == unknownStatusError
}
