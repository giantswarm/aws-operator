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

func IsKeyNotFound(err error) bool {
	aerr, ok := err.(awserr.Error)
	if ok && aerr.Code() == kms.ErrCodeNotFoundException {
		return true
	}

	return false
}
