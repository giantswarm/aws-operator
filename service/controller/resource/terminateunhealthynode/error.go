package terminateunhealthynode

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

var invalidProviderIDError = &microerror.Error{
	Kind: "invalidProviderID",
}

// IsInvalidProviderID asserts invalidConfigError.
func IsInvalidProviderID(err error) bool {
	return microerror.Cause(err) == invalidProviderIDError
}
