package tccpsubnet

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

var subnetNotFoundError = &microerror.Error{
	Kind: "subnetNotFoundError",
}

// IsSubnetNotFound asserts subnetNotFoundError.
func IsSubnetNotFound(err error) bool {
	c := microerror.Cause(err)

	aerr, ok := c.(awserr.Error)
	if ok {
		if aerr.Code() == "NoSuchEntity" {
			return true
		}
	}

	if c == subnetNotFoundError {
		return true
	}

	return false
}

var vpcNotFoundError = &microerror.Error{
	Kind: "vpcNotFoundError",
}

// IsVPCNotFound asserts vpcNotFoundError.
func IsVPCNotFound(err error) bool {
	return microerror.Cause(err) == vpcNotFoundError
}
