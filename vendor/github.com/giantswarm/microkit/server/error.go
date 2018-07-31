package server

import (
	"net/http"

	"github.com/giantswarm/microerror"
)

var invalidConfigError = &microerror.Error{
	Kind: "invalidConfigError",
}

// IsInvalidConfig asserts invalidConfigError.
func IsInvalidConfig(err error) bool {
	return microerror.Cause(err) == invalidConfigError
}

var invalidContextError = &microerror.Error{
	Kind: "invalidContextError",
}

// IsInvalidContext asserts invalidContextError.
func IsInvalidContext(err error) bool {
	return microerror.Cause(err) == invalidContextError
}

var invalidTransactionIDError = &microerror.Error{
	Kind: "invalidTransactionIDError",
}

// IsInvalidTransactionID asserts invalidTransactionIDError.
func IsInvalidTransactionID(err error) bool {
	return microerror.Cause(err) == invalidTransactionIDError
}

var serverClosedError = &microerror.Error{
	Kind: "serverClosedError",
}

// IsServerClosed asserts serverClosedError.
func IsServerClosed(err error) bool {
	if err == nil {
		return false
	}

	c := microerror.Cause(err)

	if c == http.ErrServerClosed {
		return true
	}
	if c == serverClosedError {
		return true
	}

	return false
}
