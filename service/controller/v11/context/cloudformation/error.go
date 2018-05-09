package cloudformation

import "github.com/giantswarm/microerror"

var serviceNotFound = microerror.New("cloudformation service not found")

// IsServiceNotFoundError asserts serviceNotFound.
func IsServiceNotFoundError(err error) bool {
	return microerror.Cause(err) == serviceNotFound
}
