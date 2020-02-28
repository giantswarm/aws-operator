package ensurecpcrs

import (
	"github.com/giantswarm/microerror"
)

var idSpaceExhaustedError = &microerror.Error{
	Kind: "idSpaceExhaustedError",
}

// IsIDSpaceExhausted asserts idSpaceExhaustedError.
func IsIDSpaceExhausted(err error) bool {
	return microerror.Cause(err) == idSpaceExhaustedError
}

var invalidConfigError = &microerror.Error{
	Kind: "invalidConfigError",
}

// IsInvalidConfig asserts invalidConfigError.
func IsInvalidConfig(err error) bool {
	return microerror.Cause(err) == invalidConfigError
}
