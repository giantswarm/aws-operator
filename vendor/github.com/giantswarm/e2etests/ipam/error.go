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

var notFoundError = &microerror.Error{
	Kind: "not found",
}

// IsNotFound asserts NotFoundError.
func IsNotFound(err error) bool {
	return microerror.Cause(err) == notFoundError
}

var waitError = &microerror.Error{
	Kind: "wait",
}

// IsWait asserts waitError.
func IsWait(err error) bool {
	return microerror.Cause(err) == waitError
}
