package kms

import (
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/giantswarm/microerror"
)

var invalidConfigError = microerror.New("invalid config")

// IsInvalidConfig asserts invalidConfigError.
func IsInvalidConfig(err error) bool {
	return microerror.Cause(err) == invalidConfigError
}

var keyNotFoundError = microerror.New("key not found")

// IsKeyNotFound asserts keyNotFoundError.
func IsKeyNotFound(err error) bool {
	if err == nil {
		return false
	}

	aerr, ok := err.(awserr.Error)
	if ok && aerr.Code() == kms.ErrCodeNotFoundException {
		return true
	}

	c := microerror.Cause(err)
	if c == keyNotFoundError {
		return true
	}

	return false
}
