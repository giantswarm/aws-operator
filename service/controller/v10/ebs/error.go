package ebs

import (
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/giantswarm/microerror"
)

var invalidConfigError = microerror.New("invalid config")

// IsInvalidConfig asserts invalidConfigError.
func IsInvalidConfig(err error) bool {
	return microerror.Cause(err) == invalidConfigError
}

var volumeNotFoundError = microerror.New("volume not found")

// IsVolumeNotFound asserts volume not found error from upstream's API code.
func IsVolumeNotFound(err error) bool {
	if err == nil {
		return false
	}

	c := microerror.Cause(err)
	if c == volumeNotFoundError {
		return true
	}

	aerr, ok := c.(awserr.Error)
	if !ok {
		return false
	}
	// TODO Find constant in the Go SDK for the error code.
	if aerr.Code() == "NotFound" {
		return true
	}

	return false
}
