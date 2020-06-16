package f001

import "github.com/giantswarm/microerror"

var invalidSecretError = &microerror.Error{
	Kind: "invalidSecretError",
}

func IsInvalidSecret(err error) bool {
	return microerror.Cause(err) == invalidSecretError
}
