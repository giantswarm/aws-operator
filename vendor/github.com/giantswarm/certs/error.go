package certs

import "github.com/giantswarm/microerror"

var executionError = microerror.New("execution error")

func IsExecutionError(err error) bool {
	return microerror.Cause(err) == executionError
}

var invalidConfigError = microerror.New("invalid config")

func IsInvalidConfig(err error) bool {
	return microerror.Cause(err) == invalidConfigError
}

var invalidSecretError = microerror.New("invalid secret")

func IsInvalidSecret(err error) bool {
	return microerror.Cause(err) == invalidSecretError
}

var timeoutError = microerror.New("timeout")

func IsTimeoutError(err error) bool {
	return microerror.Cause(err) == timeoutError
}

var wrongTypeError = microerror.New("wrong type")

func IsWrongTypeError(err error) bool {
	return microerror.Cause(err) == wrongTypeError
}
