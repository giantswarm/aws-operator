package credential

import "github.com/giantswarm/microerror"

var arnNotFound = microerror.New("arn not found")

// IsArnNotFoundError asserts arnNotFound.
func IsArnNotFoundError(err error) bool {
	return microerror.Cause(err) == arnNotFound
}
