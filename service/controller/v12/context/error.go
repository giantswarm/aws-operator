package context

import (
	"github.com/giantswarm/microerror"
)

var serviceNotFound = microerror.New("service not found")

// IsServiceNotFoundError asserts serviceNotFound.
func IsServiceNotFoundError(err error) bool {
	return microerror.Cause(err) == serviceNotFound
}
