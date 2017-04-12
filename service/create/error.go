package create

import (
	"github.com/juju/errgo"
)

var (
	invalidConfigError          = errgo.New("invalid config")
	secretsRetrievalFailedError = errgo.New("could not retrieve secrets")
	malformedDNSNameError       = errgo.New("could not parse DNS name in TPR")
)

// IsInvalidConfig asserts invalidConfigError.
func IsInvalidConfig(err error) bool {
	return errgo.Cause(err) == invalidConfigError
}
