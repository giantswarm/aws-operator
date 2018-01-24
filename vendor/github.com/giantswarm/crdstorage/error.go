package crdstorage

import (
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/microstorage"
	"github.com/juju/errgo"
)

var invalidConfigError = microerror.New("invalid config")

// IsInvalidConfig asserts invalidConfigError.
func IsInvalidConfig(err error) bool {
	return errgo.Cause(err) == invalidConfigError
}

// Errors from microstorage must be reused in order to match microstorage error
// matchers. This is required to fulfil the interface and pass storagetest.

var notFoundError = microstorage.NotFoundError
