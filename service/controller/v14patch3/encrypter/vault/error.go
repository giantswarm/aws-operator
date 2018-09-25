package vault

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

var keyNotFoundError = &microerror.Error{
	Kind: "keyNotFoundError",
}

// IsKeyNotFound asserts keyNotFoundError.
func IsKeyNotFound(err error) bool {
	return microerror.Cause(err) == keyNotFoundError
}

var invalidHTTPStatusCodeError = &microerror.Error{
	Kind: "invalidHTTPStatusCodeError",
}

// IsInvalidHTTPStatus asserts invalidHTTPStatusCodeError.
func IsInvalidHTTPStatus(err error) bool {
	return microerror.Cause(err) == invalidHTTPStatusCodeError
}
