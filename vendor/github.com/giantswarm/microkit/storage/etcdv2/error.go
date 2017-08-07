package etcdv2

import (
	"github.com/coreos/etcd/client"
	"github.com/giantswarm/microerror"
)

var createFailedError = microerror.New("create failed")

// IsCreateFailed asserts createFailedError.
func IsCreateFailed(err error) bool {
	return microerror.Cause(err) == createFailedError
}

// IsEtcdKeyAlreadyExists is an error matcher for the v2 etcd client.
func IsEtcdKeyAlreadyExists(err error) bool {
	if cErr, ok := err.(client.Error); ok {
		return cErr.Code == client.ErrorCodeNodeExist
	}
	return false
}

var invalidConfigError = microerror.New("invalid config")

// IsInvalidConfig asserts invalidConfigError.
func IsInvalidConfig(err error) bool {
	return microerror.Cause(err) == invalidConfigError
}

var multipleValuesError = microerror.New("multiple values")

// IsMultipleValuesFound asserts multipleValuesError.
func IsMultipleValuesFound(err error) bool {
	return microerror.Cause(err) == multipleValuesError
}

var notFoundError = microerror.New("not found")

// IsNotFound asserts notFoundError.
func IsNotFound(err error) bool {
	return microerror.Cause(err) == notFoundError
}
