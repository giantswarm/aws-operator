package apprclient

import "github.com/giantswarm/microerror"

var invalidConfigError = microerror.New("invalid config")

// IsInvalidConfig asserts invalidConfigError.
func IsInvalidConfig(err error) bool {
	return microerror.Cause(err) == invalidConfigError
}

var unknownStatusError = microerror.New("unknown status")

// IsUnknownStatus asserts unknownStatusError.
func IsUnknownStatus(err error) bool {
	return microerror.Cause(err) == unknownStatusError
}
