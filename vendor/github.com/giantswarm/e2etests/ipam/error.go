package ipam

import "github.com/giantswarm/microerror"

var alreadyExistsError = &microerror.Error{
	Kind: "alreadyExistsError",
}

// IsAlreadyExists asserts alreadyExistsError.
func IsAlreadyExists(err error) bool {
	return microerror.Cause(err) == alreadyExistsError
}

var invalidConfigError = &microerror.Error{
	Kind: "invalid config",
}

// IsInvalidConfig asserts invalidConfigError.
func IsInvalidConfig(err error) bool {
	return microerror.Cause(err) == invalidConfigError
}

var subnetsOverlapError = &microerror.Error{
	Kind: "subnetsOverlap",
}

// IssubnetsOverlap asserts subnetsOverlapError.
func IsSubnetsOverlap(err error) bool {
	return microerror.Cause(err) == subnetsOverlapError
}
