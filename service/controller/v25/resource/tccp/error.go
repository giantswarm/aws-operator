package tccp

import (
	"strings"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/giantswarm/microerror"
)

var alreadyExistsError = &microerror.Error{
	Kind: "alreadyExistsError",
}

// IsAlreadyExists asserts alreadyExistsError.
func IsAlreadyExists(err error) bool {
	c := microerror.Cause(err)

	if c == nil {
		return false
	}

	if strings.Contains(c.Error(), cloudformation.ErrCodeAlreadyExistsException) {
		return true
	}

	if c == alreadyExistsError {
		return true
	}

	return false
}

var alreadyTerminatedError = &microerror.Error{
	Kind: "alreadyTerminatedError",
}

// IsAlreadyTerminated asserts alreadyTerminatedError. Here we also check for
// the AWS error code IncorrectInstanceState. The AWS errors might look like the
// following example.
//
//     IncorrectInstanceState: The instance 'i-0b26c88f3546aefee' must be in a 'running', 'pending', 'stopping' or 'stopped' state for this operation.
//
func IsAlreadyTerminated(err error) bool {
	c := microerror.Cause(err)

	aerr, ok := c.(awserr.Error)
	if ok {
		if aerr.Code() == "IncorrectInstanceState" {
			return true
		}
	}

	if c == alreadyTerminatedError {
		return true
	}

	return false
}

var deleteInProgressError = &microerror.Error{
	Kind: "deleteInProgressError",
}

// IsDeleteInProgress asserts deleteInProgressError.
func IsDeleteInProgress(err error) bool {
	c := microerror.Cause(err)

	if c == nil {
		return false
	}

	if strings.Contains(c.Error(), cloudformation.ResourceStatusDeleteInProgress) {
		return true
	}

	if c == deleteInProgressError {
		return true
	}

	return false
}

var deletionMustBeRetriedError = &microerror.Error{
	Kind: "deletionMustBeRetriedError",
}

// IsDeletionMustBeRetried asserts deletionMustBeRetriedError.
func IsDeletionMustBeRetried(err error) bool {
	return microerror.Cause(err) == deletionMustBeRetriedError
}

var executionFailedError = &microerror.Error{
	Kind: "executionFailedError",
}

// IsExecutionFailed asserts executionFailedError.
func IsExecutionFailed(err error) bool {
	return microerror.Cause(err) == executionFailedError
}

var invalidConfigError = &microerror.Error{
	Kind: "invalidConfigError",
}

// IsInvalidConfig asserts invalidConfigError.
func IsInvalidConfig(err error) bool {
	return microerror.Cause(err) == invalidConfigError
}

var notExistsError = &microerror.Error{
	Kind: "notExistsError",
}

// IsNotExists asserts notExistsError.
func IsNotExists(err error) bool {
	c := microerror.Cause(err)

	if c == nil {
		return false
	}

	if strings.Contains(c.Error(), "does not exist") {
		return true
	}

	if c == notExistsError {
		return true
	}

	return false
}

var resourceNotReadyError = &microerror.Error{
	Kind: "resourceNotReadyError",
}

// IsResourceNotReady asserts resourceNotReadyError.
func IsResourceNotReady(err error) bool {
	c := microerror.Cause(err)

	if c == nil {
		return false
	}

	if strings.Contains(c.Error(), "ResourceNotReady") {
		return true
	}

	if c == resourceNotReadyError {
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
