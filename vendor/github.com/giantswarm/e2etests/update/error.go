package update

import "github.com/giantswarm/microerror"

var invalidConfigError = &microerror.Error{
	Kind: "invalidConfigError",
}

// IsInvalidConfig asserts invalidConfigError.
func IsInvalidConfig(err error) bool {
	return microerror.Cause(err) == invalidConfigError
}

var missesDesiredStatusError = &microerror.Error{
	Kind: "missesDesiredStatusError",
}

// IsMissesDesiredStatus asserts missesDesiredStatusError.
func IsMissesDesiredStatus(err error) bool {
	return microerror.Cause(err) == missesDesiredStatusError
}
