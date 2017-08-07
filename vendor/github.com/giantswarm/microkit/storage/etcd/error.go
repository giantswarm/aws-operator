package etcd

import (
	"github.com/giantswarm/microerror"
)

var createFailedError = microerror.New("create failed")

// IsCreateFailed asserts createFailedError.
func IsCreateFailed(err error) bool {
	return microerror.Cause(err) == createFailedError
}

var invalidConfigError = microerror.New("invalid config")

// IsInvalidConfig asserts invalidConfigError.
func IsInvalidConfig(err error) bool {
	return microerror.Cause(err) == invalidConfigError
}

var multipleValuesError = microerror.New("multiple values")

// IsMultipleValuesFound asserts multipleValuesError.
func IsMultipleValuesFound(err error) bool {
	return microerror.Cause(err) == multipleValuesError
}

var notFoundError = microerror.New("not found")

// IsNotFound asserts notFoundError.
func IsNotFound(err error) bool {
	return microerror.Cause(err) == notFoundError
}
