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
