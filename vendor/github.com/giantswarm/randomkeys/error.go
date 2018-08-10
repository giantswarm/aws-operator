package randomkeys

import "github.com/giantswarm/microerror"

var executionError = &microerror.Error{
	Kind: "executionError",
}

func IsExecutionError(err error) bool {
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

func IsTimeoutError(err error) bool {
	return microerror.Cause(err) == timeoutError
}

var wrongTypeError = &microerror.Error{
	Kind: "wrongTypeError",
}

func IsWrongTypeError(err error) bool {
	return microerror.Cause(err) == wrongTypeError
}
