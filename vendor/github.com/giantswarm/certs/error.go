package certs

import "github.com/giantswarm/microerror"

var executionError = &microerror.Error{
	Kind: "executionError",
}

func IsExecution(err error) bool {
	return microerror.Cause(err) == executionError
}

var invalidConfigError = &microerror.Error{
	Kind: "invalidConfigError",
}

func IsInvalidConfig(err error) bool {
	return microerror.Cause(err) == invalidConfigError
}

var invalidSecretError = &microerror.Error{
	Kind: "invalidSecretError",
}

func IsInvalidSecret(err error) bool {
	return microerror.Cause(err) == invalidSecretError
}

var timeoutError = &microerror.Error{
	Kind: "timeoutError",
}

func IsTimeout(err error) bool {
	return microerror.Cause(err) == timeoutError
}

var wrongTypeError = &microerror.Error{
	Kind: "wrongTypeError",
}

func IsWrongType(err error) bool {
	return microerror.Cause(err) == wrongTypeError
}
