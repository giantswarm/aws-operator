package statusresource

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

var missingLabelError = &microerror.Error{
	Kind: "missingLabelError",
}

// IsMissingLabel asserts missingLabelError.
func IsMissingLabel(err error) bool {
	return microerror.Cause(err) == missingLabelError
}

var notFoundError = &microerror.Error{
	Kind: "notFoundError",
}

// IsNotFound asserts notFoundError.
func IsNotFound(err error) bool {
	return microerror.Cause(err) == notFoundError
}
