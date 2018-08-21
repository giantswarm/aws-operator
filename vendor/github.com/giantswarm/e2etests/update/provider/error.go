package provider

import "github.com/giantswarm/microerror"

var invalidConfigError = &microerror.Error{
	Kind: "invalidConfigError",
}

// IsInvalidConfig asserts invalidConfigError.
func IsInvalidConfig(err error) bool {
	return microerror.Cause(err) == invalidConfigError
}

var versionNotFoundError = &microerror.Error{
	Kind: "versionNotFoundError",
}

// IsVersionNotFound asserts versionNotFoundError.
func IsVersionNotFound(err error) bool {
	return microerror.Cause(err) == versionNotFoundError
}
