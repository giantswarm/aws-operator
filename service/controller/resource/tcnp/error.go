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
//     https://github.com/giantswarm/fmt/blob/master/go/errors.md#matching-errors
//
var executionFailedError = &microerror.Error{
	Kind: "executionFailedError",
}

// executionCF... is an error type for situations where Resource execution
// cannot continue and must always fall back to operatorkit.
//
// Operatorkit creates an error event on the corresponding Kubernetes objects.
var executionCFCreateError = &microerror.Error{
	Kind: "CFCreateFailed",
	Desc: fmt.Sprintf("The tenant cluster's node pool cloud formation stack has stack status %#q", cloudformation.StackStatusCreateFailed),
}
var executionCFUpdateRollbackError = &microerror.Error{
	Kind: "CFUpdateRollbackFailed",
	Desc: fmt.Sprintf("The tenant cluster's node pool cloud formation stack has stack status %#q", cloudformation.StackStatusUpdateRollbackFailed),
}

var executionCFRollbackError = &microerror.Error{
	Kind: "CFRollbackFailed",
	Desc: fmt.Sprintf("The tenant cluster's node pool cloud formation stack has stack status %#q", cloudformation.StackStatusRollbackFailed),
}

var executionCFDeleteError = &microerror.Error{
	Kind: "CFDeleteFailed",
	Desc: fmt.Sprintf("The tenant cluster's node pool cloud formation stack has stack status %#q", cloudformation.StackStatusDeleteFailed),
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
