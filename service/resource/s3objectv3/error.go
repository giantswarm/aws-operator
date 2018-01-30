package s3objectv3

import (
	"strings"

	"github.com/giantswarm/microerror"
)

var invalidConfigError = microerror.New("invalid config")

// IsInvalidConfig asserts invalidConfigError.
func IsInvalidConfig(err error) bool {
	return microerror.Cause(err) == invalidConfigError
}

var wrongTypeError = microerror.New("wrong type")

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
