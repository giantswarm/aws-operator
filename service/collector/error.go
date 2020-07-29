package collector

import (
	"github.com/giantswarm/microerror"
)

var invalidConfigError = &microerror.Error{
	Kind: "invalidConfigError",
}

// IsInvalidConfig asserts invalidConfigError.
func IsInvalidConfig(err error) bool {
	return microerror.Cause(err) == invalidConfigError
}

var invalidResourceError = &microerror.Error{
	Kind: "invalidResourceError",
}

// IsInvalidResource asserts invalidResourceError.
func IsInvalidResource(err error) bool {
	return microerror.Cause(err) == invalidResourceError
}

var nilLimitError = &microerror.Error{
	Kind: "nilLimitError",
}

// IsNilLimit asserts nilLimitError.
func IsNilLimit(err error) bool {
	return microerror.Cause(err) == nilLimitError
}

var nilUsageError = &microerror.Error{
	Kind: "nilUsageError",
}

// IsNilUsage asserts nilUsageError.
func IsNilUsage(err error) bool {
	return microerror.Cause(err) == nilUsageError
}
