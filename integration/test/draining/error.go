package draining

import (
	"github.com/giantswarm/microerror"
)

var e2eAppError = &microerror.Error{
	Kind: "e2eAppError",
}

// IsE2EAppError asserts e2eAppError.
func IsE2EAppError(err error) bool {
	return microerror.Cause(err) == e2eAppError
}

var podError = &microerror.Error{
	Kind: "podError",
}

// IsPod asserts podError.
func IsPod(err error) bool {
	return microerror.Cause(err) == podError
}
