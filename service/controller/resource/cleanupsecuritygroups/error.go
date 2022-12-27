package cleanupsecuritygroups

import (
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/giantswarm/microerror"
)

var dependencyViolationError = &microerror.Error{
	Kind: "dependencyViolationError",
}

// IsDependencyViolation asserts dependencyViolationError. Additionally it
// asserts AWS errors which may look like the following.
//
//	DependencyViolation: resource sg-07423aeb02946f323 has a dependent object\n\tstatus code: 400, request id: c16da859-433c-4e59-b598-ef17f9faa770
func IsDependencyViolation(err error) bool {
	c := microerror.Cause(err)

	if c == dependencyViolationError {
		return true
	}

	aerr, ok := c.(awserr.Error)
	if !ok {
		return false
	}
	if aerr.Code() == "DependencyViolation" {
		return true
	}

	return false
}

var invalidConfigError = &microerror.Error{
	Kind: "invalidConfigError",
}

// IsInvalidConfig asserts invalidConfigError.
func IsInvalidConfig(err error) bool {
	return microerror.Cause(err) == invalidConfigError
}
