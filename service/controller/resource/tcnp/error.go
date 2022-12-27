package tcnp

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/giantswarm/microerror"
)

// executionFailedError is an error type for situations where Resource execution
// cannot continue and must always fall back to operatorkit.
//
// This error should never be matched against and therefore there is no matcher
// implement. For further information see:
//
//	https://github.com/giantswarm/fmt/blob/master/go/errors.md#matching-errors
var executionFailedError = &microerror.Error{
	Kind: "executionFailedError",
}

// event... is an error type for situations where we want to create an Kubernetes event in operatorkit.
var eventCFCreateError = &microerror.Error{
	Kind: "CFCreateFailed",
	Desc: fmt.Sprintf("The tenant cluster's node pool cloud formation stack has stack status %#q", cloudformation.StackStatusCreateFailed),
}
var eventCFUpdateRollbackError = &microerror.Error{
	Kind: "CFUpdateRollbackFailed",
	Desc: fmt.Sprintf("The tenant cluster's node pool cloud formation stack has stack status %#q", cloudformation.StackStatusUpdateRollbackFailed),
}

var eventCFRollbackError = &microerror.Error{
	Kind: "CFRollbackFailed",
	Desc: fmt.Sprintf("The tenant cluster's node pool cloud formation stack has stack status %#q", cloudformation.StackStatusRollbackFailed),
}

var eventCFDeleteError = &microerror.Error{
	Kind: "CFDeleteFailed",
	Desc: fmt.Sprintf("The tenant cluster's node pool cloud formation stack has stack status %#q", cloudformation.StackStatusDeleteFailed),
}

// IsDeleteFailed asserts eventCFDeleteError.
func IsDeleteFailed(err error) bool {
	c := microerror.Cause(err)

	if c == nil {
		return false
	}

	if strings.Contains(c.Error(), cloudformation.ResourceStatusDeleteFailed) {
		return true
	}

	if c == eventCFDeleteError {
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

var outputNotFoundError = &microerror.Error{
	Kind: "outputNotFoundError",
}

// IsOutputNotFound asserts outputNotFoundError.
func IsOutputNotFound(err error) bool {
	return microerror.Cause(err) == outputNotFoundError
}

var tccpnNotUpdatedError = &microerror.Error{
	Kind: "tccpnNotUpdatedError",
}

// IsTccpnNotUpdated asserts timeoutError.
func IsTccpnNotUpdated(err error) bool {
	return microerror.Cause(err) == tccpnNotUpdatedError
}
