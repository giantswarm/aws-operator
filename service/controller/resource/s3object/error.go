package s3object

import (
	"strings"

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

var wrongTypeError = &microerror.Error{
	Kind: "wrongTypeError",
}

// IsWrongType asserts wrongTypeError.
func IsWrongType(err error) bool {
	return microerror.Cause(err) == wrongTypeError
}

// IsObjectNotFound asserts object not found error from upstream's API message.
func IsObjectNotFound(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(microerror.Cause(err).Error(), "NoSuchKey: The specified key does not exist")
}

// IsBucketNotFound asserts object not found error from upstream's API message.
func IsBucketNotFound(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(microerror.Cause(err).Error(), "NoSuchBucket: The specified bucket does not exist")
}

// IsKeyNotFound asserts key not found error from upstream's API message.
func IsKeyNotFound(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(microerror.Cause(err).Error(), "NotFoundException: Alias arn:aws:kms:")
}
