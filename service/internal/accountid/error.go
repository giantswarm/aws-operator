package accountid

import "github.com/giantswarm/microerror"

var invalidAccountIDError = &microerror.Error{
	Kind: "invalidAccountIDError",
}

// IsInvalidAccountID asserts invalidAccountIDError.
func IsInvalidAccountID(err error) bool {
	return microerror.Cause(err) == invalidAccountIDError
}

var invalidConfigError = &microerror.Error{
	Kind: "invalidConfigError",
}

// IsInvalidConfig asserts invalidConfigError.
func IsInvalidConfig(err error) bool {
	return microerror.Cause(err) == invalidConfigError
}
