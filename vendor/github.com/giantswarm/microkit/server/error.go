package server

import (
	"github.com/giantswarm/microerror"
	"github.com/juju/errgo"
)

func errorTrace(err error) string {
	switch e := err.(type) {
	case *errgo.Err:
		return e.GoString()
	}
	return "n/a"
}

var invalidConfigError = microerror.New("invalid config")

// IsInvalidConfig asserts invalidConfigError.
func IsInvalidConfig(err error) bool {
	return microerror.Cause(err) == invalidConfigError
}

var invalidContextError = microerror.New("invalid context")

// IsInvalidContext asserts invalidContextError.
func IsInvalidContext(err error) bool {
	return microerror.Cause(err) == invalidContextError
}

var invalidTransactionIDError = microerror.New("invalid transaction ID")

// IsInvalidTransactionID asserts invalidTransactionIDError.
func IsInvalidTransactionID(err error) bool {
	return microerror.Cause(err) == invalidTransactionIDError
}
