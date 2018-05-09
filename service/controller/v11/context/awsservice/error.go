package awsservice

import "github.com/giantswarm/microerror"

var serviceNotFound = microerror.New("aws service not found")

// IsServiceNotFoundError asserts serviceNotFound.
func IsServiceNotFoundError(err error) bool {
	return microerror.Cause(err) == serviceNotFound
}
