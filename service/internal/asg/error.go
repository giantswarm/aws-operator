package asg

import "github.com/giantswarm/microerror"

var invalidConfigError = &microerror.Error{
	Kind: "invalidConfigError",
}

// IsInvalidConfig asserts invalidConfigError.
func IsInvalidConfig(err error) bool {
	return microerror.Cause(err) == invalidConfigError
}

var noASGError = &microerror.Error{
	Kind: "noASGError",
}

// IsNoASG asserts noASGError.
func IsNoASG(err error) bool {
	return microerror.Cause(err) == noASGError
}

var noDrainableError = &microerror.Error{
	Kind: "noDrainableError",
}

// IsNoDrainable asserts noDrainableError.
func IsNoDrainable(err error) bool {
	return microerror.Cause(err) == noDrainableError
}
