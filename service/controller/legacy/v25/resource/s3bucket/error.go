package s3bucket

import (
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/giantswarm/microerror"
)

var invalidConfigError = &microerror.Error{
	Kind: "invalidConfigError",
}

// IsInvalidConfig asserts invalidConfigError.
func IsInvalidConfig(err error) bool {
	return microerror.Cause(err) == invalidConfigError
}

var notFoundError = &microerror.Error{
	Kind: "notFoundError",
}

// IsNotFound asserts notFoundError.
func IsNotFound(err error) bool {
	return microerror.Cause(err) == notFoundError
}

var bucketNotFoundError = &microerror.Error{
	Kind: "bucketNotFoundError",
}

// IsBucketNotFound asserts bucket not found error from upstream's API code.
func IsBucketNotFound(err error) bool {
	c := microerror.Cause(err)
	aerr, ok := c.(awserr.Error)
	if !ok {
		return false
	}
	// hack for HeadBucket request that returns a wrong error code
	if aerr.Code() == "NotFound" {
		return true
	}
	if aerr.Code() == s3.ErrCodeNoSuchBucket {
		return true
	}
	if c == bucketNotFoundError {
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

var bucketNotEmptyError = &microerror.Error{
	Kind: "bucketNotEmptyError",
}

// IsBucketNotEmpty asserts bucketNotEmptyError. It also checks for
// BucketNotEmpty error codes from the AWS SDK. An error we expect looks like
// the one below.
//
//     BucketNotEmpty: The bucket you tried to delete is not empty\n\tstatus code: 409, request id: 4B2CDF3222517C9D, host id: mOJAOuJsV/3CEeAkyTw1k3s5HLFsa5PHMkUfZv5lqtOKxiR67jclbqIHrzvtDa7E676h908MIY0=
//
func IsBucketNotEmpty(err error) bool {
	c := microerror.Cause(err)
	aerr, ok := c.(awserr.Error)
	if !ok {
		return false
	}

	if aerr.Code() == "BucketNotEmpty" {
		return true
	}
	if c == bucketNotEmptyError {
		return true
	}

	return false
}

var wrongTypeError = &microerror.Error{
	Kind: "wrongTypeError",
}

// IsWrongType asserts wrongTypeError.
func IsWrongType(err error) bool {
	return microerror.Cause(err) == wrongTypeError
}
