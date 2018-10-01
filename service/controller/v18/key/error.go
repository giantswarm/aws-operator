package key

import "github.com/giantswarm/microerror"

var wrongTypeError = &microerror.Error{
	Kind: "wrongTypeError",
}

// IsWrongTypeError asserts wrongTypeError.
func IsWrongTypeError(err error) bool {
	return microerror.Cause(err) == wrongTypeError
}

var malformedCloudConfigKeyError = &microerror.Error{
	Kind: "malformedCloudConfigKeyError",
}

// IsMalformedCloudConfigKey asserts malformedCloudConfigKeyError.
func IsMalformedCloudConfigKey(err error) bool {
	return microerror.Cause(err) == malformedCloudConfigKeyError
}

var missingCloudConfigKeyError = &microerror.Error{
	Kind: "missingCloudConfigKeyError",
}

// IsMissingCloudConfigKey asserts missingCloudConfigKeyError.
func IsMissingCloudConfigKey(err error) bool {
	return microerror.Cause(err) == missingCloudConfigKeyError
}

var notFoundError = &microerror.Error{
	Kind: "notFoundError",
}

// IsNotFound asserts notFoundError.
func IsNotFound(err error) bool {
	return microerror.Cause(err) == notFoundError
}

var invalidConfigError = &microerror.Error{
	Kind: "invalidConfigError",
}

// IsInvalidConfig asserts invalidConfigError.
func IsInvalidConfig(err error) bool {
	return microerror.Cause(err) == invalidConfigError
}
