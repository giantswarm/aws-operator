package s3bucket

import (
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/giantswarm/microerror"
)

var invalidConfigError = microerror.New("invalid config")

// IsInvalidConfig asserts invalidConfigError.
func IsInvalidConfig(err error) bool {
	return microerror.Cause(err) == invalidConfigError
}

var notFoundError = microerror.New("not found")

// IsNotFound asserts notFoundError.
func IsNotFound(err error) bool {
	return microerror.Cause(err) == notFoundError
}

var noSuchBucketError = microerror.New("Not Found")

// IsBucketNotFound asserts bucket not found error from upstream's API code.
func IsBucketNotFound(err error) bool {
	aerr, ok := err.(awserr.Error)
	if !ok {
		return false
	}
	if microerror.Cause(aerr) == noSuchBucketError {
		return true
	}

	return false
}

// IsBucketAlreadyExists asserts bucket already exists error from upstream's
// API code.
func IsBucketAlreadyExists(err error) bool {
	aerr, ok := err.(awserr.Error)
	if !ok {
		return false
	}
	if aerr.Code() == s3.ErrCodeBucketAlreadyExists {
		return true
	}

	return false
}

// IsBucketAlreadyOwnedByYou asserts bucket already owned by you error from
// upstream's API code.
func IsBucketAlreadyOwnedByYou(err error) bool {
	aerr, ok := err.(awserr.Error)
	if !ok {
		return false
	}
	if aerr.Code() == s3.ErrCodeBucketAlreadyOwnedByYou {
		return true
	}

	return false
}

var wrongTypeError = microerror.New("wrong type")

// IsWrongType asserts wrongTypeError.
func IsWrongType(err error) bool {
	return microerror.Cause(err) == wrongTypeError
}
