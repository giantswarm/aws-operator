package cloudconfig

import "github.com/giantswarm/microerror"

var executionFailedError = &microerror.Error{
	Kind: "executionFailedError",
}

var invalidConfigError = &microerror.Error{
	Kind: "invalidConfigError",
}

// IsInvalidConfig asserts invalidConfigError.
func IsInvalidConfig(err error) bool {
	return microerror.Cause(err) == invalidConfigError
}

var notFoundError = &microerror.Error{
	Kind: "notFoundError",
}

// IsNotFound asserts notFoundError.
func IsNotFound(err error) bool {
	return microerror.Cause(err) == notFoundError
}

var timeoutError = &microerror.Error{
	Kind: "timeoutError",
}

// IsTimeout asserts timeoutError.
func IsTimeout(err error) bool {
	return microerror.Cause(err) == timeoutError
}
