package s3bucket

import (
	"strings"

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

// IsBucketNotFound asserts bucket not found error from upstream's API code.
func IsBucketNotFound(err error) bool {
	aerr, ok := err.(awserr.Error)
	if !ok {
		return false
	}
	if aerr.Code() == s3.ErrCodeNoSuchBucket {
		return true
	}
	// AWS returns response Not Found with a wrong format, so
	// for now a hack to detect bucket does not exist
	if strings.Contains(aerr.Code(), "Not Found") {
		return true
	}
	if microerror.Cause(err) == notFoundError {
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
