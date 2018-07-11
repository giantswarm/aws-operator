package vault

import (
	"github.com/giantswarm/microerror"
)

var invalidConfigError = microerror.New("invalid config")

// IsInvalidConfig asserts invalidConfigError.
func IsInvalidConfig(err error) bool {
	return microerror.Cause(err) == invalidConfigError
}

var keyNotFoundError = microerror.New("key not found")

// IsKeyNotFound asserts keyNotFoundError.
func IsKeyNotFound(err error) bool {
	return microerror.Cause(err) == keyNotFoundError
}

var invalidHTTPStatusCodeError = microerror.New("invalid HTTP status code")

// IsInvalidHTTPStatus asserts invalidHTTPStatusCodeError.
func IsInvalidHTTPStatus(err error) bool {
	return microerror.Cause(err) == invalidHTTPStatusCodeError
}
