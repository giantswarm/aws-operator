package cleanupenis

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

var invalidENIStatusError = &microerror.Error{
	Kind: "invalidENIStatusError",
}

// IsInvalidENIStatus asserts invalidENIStatusError.
func IsInvalidENIStatus(err error) bool {
	return microerror.Cause(err) == invalidENIStatusError
}
