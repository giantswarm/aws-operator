package ebs

import (
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/giantswarm/microerror"
)

// executionFailedError is an error type for situations where Resource execution
// cannot continue and must always fall back to operatorkit.
//
// This error should never be matched against and therefore there is no matcher
// implement. For further information see:
//
//	https://github.com/giantswarm/fmt/blob/master/go/errors.md#matching-errors
var executionFailedError = &microerror.Error{
	Kind: "executionFailedError",
}

var invalidConfigError = &microerror.Error{
	Kind: "invalidConfigError",
}

// IsInvalidConfig asserts invalidConfigError.
func IsInvalidConfig(err error) bool {
	return microerror.Cause(err) == invalidConfigError
}

var volumeNotFoundError = &microerror.Error{
	Kind: "volumeNotFoundError",
}

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

var volumeAttachedError = &microerror.Error{
	Kind: "volumeAttachedError",
}

// IsVolumeAttached asserts volumeAttachedError.
func IsVolumeAttached(err error) bool {
	return microerror.Cause(err) == volumeAttachedError
}
