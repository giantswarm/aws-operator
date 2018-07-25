package statusresource

import (
	"github.com/giantswarm/microerror"
)

var invalidConfigError = microerror.New("invalid config")

// IsInvalidConfig asserts invalidConfigError.
func IsInvalidConfig(err error) bool {
	return microerror.Cause(err) == invalidConfigError
}

var missingLabelError = microerror.New("missing label")

// IsMissingLabel asserts missingLabelError.
func IsMissingLabel(err error) bool {
	return microerror.Cause(err) == missingLabelError
}
