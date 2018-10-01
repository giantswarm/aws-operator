package hostedzone

import "github.com/giantswarm/microerror"

var invalidConfigError = &microerror.Error{
	Kind: "invalidConfigError",
}

// IsInvalidConfig asserts invalidConfigError.
func IsInvalidConfig(err error) bool {
	return microerror.Cause(err) == invalidConfigError
}

var hostedZoneNotFoundError = &microerror.Error{
	Kind: "hostedZoneNotFoundError",
}

// IsHostedZoneNotFound asserts hostedZoneNotFoundError.
func IsHostedZoneNotFound(err error) bool {
	return microerror.Cause(err) == hostedZoneNotFoundError
}
