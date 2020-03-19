package v_4_9_1

import "github.com/giantswarm/microerror"

var componentNotFoundError = &microerror.Error{
	Kind: "componentNotFound",
}

// IsComponentNotFound asserts componentNotFoundError.
func IsComponentNotFound(err error) bool {
	return microerror.Cause(err) == componentNotFoundError
}

var invalidConfigError = &microerror.Error{
	Kind: "invalidConfigError",
}

// IsInvalidConfig asserts invalidConfigError.
func IsInvalidConfig(err error) bool {
	return microerror.Cause(err) == invalidConfigError
}
