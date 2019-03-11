package peerrolearn

import (
	"github.com/aws/aws-sdk-go/aws/awserr"
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
	c := microerror.Cause(err)

	aerr, ok := c.(awserr.Error)
	if ok {
		if aerr.Code() == "NoSuchEntity" {
			return true
		}
	}

	if c == notFoundError {
		return true
	}

	return false
}
