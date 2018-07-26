package migration

import (
	"github.com/giantswarm/microerror"
)

var invalidConfigError = microerror.New("invalid config")

// IsInvalidConfig asserts invalidConfigError.
func IsInvalidConfig(err error) bool {
	return microerror.Cause(err) == invalidConfigError
}

var malformedDomainError = microerror.New("malformed domain")

// IsMalformedDomain asserts malformedDomainError.
func IsMalformedDomain(err error) bool {
	return microerror.Cause(err) == malformedDomainError
}
