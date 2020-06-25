package tccpazs

import (
	"github.com/giantswarm/microerror"
	"k8s.io/apimachinery/pkg/api/errors"
)

var invalidConfigError = &microerror.Error{
	Kind: "invalid config",
}

// IsInvalidConfig asserts invalidConfigError.
func IsInvalidConfig(err error) bool {
	return microerror.Cause(err) == invalidConfigError
}

var notFoundError = &microerror.Error{
	Kind: "notFoundError",
}

// IsNotFound asserts notFoundError.
func IsNotFound(err error) bool {
	c := microerror.Cause(err)
	return c == notFoundError || errors.IsNotFound(c)
}

var tooManyCRsError = &microerror.Error{
	Kind: "tooManyCRsError",
	Desc: "There is only a single G8sControlPlane CR allowed with the current implementation.",
}

// IsTooManyCRsError asserts tooManyCRsError.
func IsTooManyCRsError(err error) bool {
	return microerror.Cause(err) == tooManyCRsError
}
