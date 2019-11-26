package tccp

import (
	"strings"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/giantswarm/microerror"
)

// executionFailedError is an error type for situations where Resource execution
// cannot continue and must always fall back to operatorkit.
//
// This error should never be matched against and therefore there is no matcher
// implement. For further information see:
//
//     https://github.com/giantswarm/fmt/blob/master/go/errors.md#matching-errors
//
var executionFailedError = &microerror.Error{
	Kind: "executionFailedError",
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

var updateInProgressError = &microerror.Error{
	Kind: "updateInProgressError",
}

// IsUpdateInProgress asserts updateInProgressError.
func IsUpdateInProgress(err error) bool {
	c := microerror.Cause(err)

	if c == nil {
		return false
	}

	if strings.Contains(c.Error(), cloudformation.ResourceStatusUpdateInProgress) {
		return true
	}

	if c == updateInProgressError {
		return true
	}

	return false
}

var emptyAmazonAccountIDError = &microerror.Error{
	Kind: "emptyAmazonAccountIDError",
}

// IsEmptyAmazonAccountID asserts emptyAmazonAccountIDError.
func IsEmptyAmazonAccountID(err error) bool {
	return microerror.Cause(err) == emptyAmazonAccountIDError
}

var malformedAmazonAccountIDError = &microerror.Error{
	Kind: "malformedAmazonAccountIDError",
}

// IsMalformedAmazonAccountID asserts malformedAmazonAccountIDError.
func IsMalformedAmazonAccountID(err error) bool {
	return microerror.Cause(err) == malformedAmazonAccountIDError
}

var notFoundError = &microerror.Error{
	Kind: "notFoundError",
}

// IsNotFound asserts notFoundError
func IsNotFound(err error) bool {
	return microerror.Cause(err) == notFoundError
}

var wrongAmazonAccountIDLengthError = &microerror.Error{
	Kind: "wrongAmazonAccountIDLengthError",
}

// IsWrongAmazonAccountIDLength asserts wrongAmazonAccountIDLengthError.
func IsWrongAmazonAccountIDLength(err error) bool {
	return microerror.Cause(err) == wrongAmazonAccountIDLengthError
}

var wrongTypeError = &microerror.Error{
	Kind: "wrongTypeError",
}

// IsWrongType asserts wrongTypeError.
func IsWrongType(err error) bool {
	return microerror.Cause(err) == wrongTypeError
}

var tooFewResultsError = &microerror.Error{
	Kind: "tooFewResultsError",
}

// IsTooFewResults asserts tooFewResultsError.
func IsTooFewResults(err error) bool {
	return microerror.Cause(err) == tooFewResultsError
}

var tooManyResultsError = &microerror.Error{
	Kind: "tooManyResultsError",
}

// IsTooManyResults asserts tooManyResultsError.
func IsTooManyResults(err error) bool {
	return microerror.Cause(err) == tooManyResultsError
}
