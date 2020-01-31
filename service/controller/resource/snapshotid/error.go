package snapshotid

import (
	"github.com/giantswarm/microerror"
)

var invalidConfigError = &microerror.Error{
	Kind: "invalidConfigError",
}

var notExistsError = &microerror.Error{
	Kind: "notExistsError",
}

var executionFailedError = &microerror.Error{
	Kind: "executionFailedError",
}

// IsInvalidConfig asserts invalidConfigError.
func IsInvalidConfig(err error) bool {
	return microerror.Cause(err) == invalidConfigError
}
