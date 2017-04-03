package create

import (
	"github.com/juju/errgo"
)

var (
	invalidConfigError          = errgo.New("invalid config")
	secretsRetrievalFailedError = errgo.New("could not retrieve secrets")
)

// IsInvalidConfig asserts invalidConfigError.
func IsInvalidConfig(err error) bool {
	return errgo.Cause(err) == invalidConfigError
}
