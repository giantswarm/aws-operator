package credential

import "github.com/giantswarm/microerror"

var arnNotFound = &microerror.Error{
	Kind: "arnNotFound",
}

// IsArnNotFoundError asserts arnNotFound.
func IsArnNotFoundError(err error) bool {
	return microerror.Cause(err) == arnNotFound
}
