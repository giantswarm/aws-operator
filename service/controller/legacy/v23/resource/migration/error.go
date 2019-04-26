package migration

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

var malformedDomainError = &microerror.Error{
	Kind: "malformedDomainError",
}

// IsMalformedDomain asserts malformedDomainError.
func IsMalformedDomain(err error) bool {
	return microerror.Cause(err) == malformedDomainError
}
