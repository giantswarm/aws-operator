package cloudformation

import (
	"strings"

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

var notFoundError = &microerror.Error{
	Kind: "notFoundError",
}

// IsNotFound asserts notFoundError.
func IsNotFound(err error) bool {
	return microerror.Cause(err) == notFoundError
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
