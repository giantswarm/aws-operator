package legacy

import (
	"strings"

	"github.com/giantswarm/microerror"
)

const (
	// Format for masked idleTimeoutSecondsOutOfRangeError.
	idleTimeoutSecondsOutOfRangeErrorFormat string = "ELB idle timeout seconds %s cannot exceed AWS maximum of 3600"
)

var idleTimeoutSecondsOutOfRangeError = &microerror.Error{
	Kind: "idleTimeoutSecondsOutOfRangeError",
}

// IsIdleTimeoutSecondsOutOfRangeError asserts idleTimeoutSecondsOutOfRangeError.
func IsIdleTimeoutSecondsOutOfRangeError(err error) bool {
	return microerror.Cause(err) == idleTimeoutSecondsOutOfRangeError
}

var invalidConfigError = &microerror.Error{
	Kind: "invalidConfigError",
}

// IsInvalidConfig asserts invalidConfigError.
func IsInvalidConfig(err error) bool {
	return microerror.Cause(err) == invalidConfigError
}

var invalidCloudconfigExtensionNameError = &microerror.Error{
	Kind: "invalidCloudconfigExtensionNameError",
}

//  asserts invalidCloudconfigExtensionNameError.
func IsInvalidCloudconfigExtensionName(err error) bool {
	return microerror.Cause(err) == invalidCloudconfigExtensionNameError
}

var malformedCloudConfigKeyError = &microerror.Error{
	Kind: "malformedCloudConfigKeyError",
}

// IsMalformedCloudConfigKey asserts malformedCloudConfigKeyError.
func IsMalformedCloudConfigKey(err error) bool {
	return microerror.Cause(err) == malformedCloudConfigKeyError
}

var missingCloudConfigKeyError = &microerror.Error{
	Kind: "missingCloudConfigKeyError",
}

// IsMissingCloudConfigKey asserts missingCloudConfigKeyError.
func IsMissingCloudConfigKey(err error) bool {
	return microerror.Cause(err) == missingCloudConfigKeyError
}

var secretsRetrievalFailedError = &microerror.Error{
	Kind: "secretsRetrievalFailedError",
}

// IsSecretsRetrievalFailed asserts secretsRetrievalFailedError.
func IsSecretsRetrievalFailed(err error) bool {
	return microerror.Cause(err) == secretsRetrievalFailedError
}

// Validation errors

var invalidAvailabilityZoneError = &microerror.Error{
	Kind: "invalidAvailabilityZoneError",
}

// IsInvalidAvailabilityZone asserts invalidAvailabilityZoneError.
func IsInvalidAvailabilityZone(err error) bool {
	return microerror.Cause(err) == invalidAvailabilityZoneError
}

var workersListEmptyError = &microerror.Error{
	Kind: "workersListEmptyError",
}

// IsWorkersListEmpty asserts workersListEmptyError.
func IsWorkersListEmpty(err error) bool {
	return microerror.Cause(err) == workersListEmptyError
}

var differentImageIDsError = &microerror.Error{
	Kind: "differentImageIDsError",
}

// IsDifferentImageIDs assert differentImageIDsError.
func IsDifferentImageIDs(err error) bool {
	return microerror.Cause(err) == differentImageIDsError
}

var differentInstanceTypesError = &microerror.Error{
	Kind: "differentInstanceTypesError",
}

// IsDifferentInstanceTypes asserts differentInstanceTypesError.
func IsDifferentInstanceTypes(err error) bool {
	return microerror.Cause(err) == differentInstanceTypesError
}

var invalidMasterNodeCountError = &microerror.Error{
	Kind: "invalidMasterNodeCountError",
}

// IsInvalidMasterNodeCount asserts invalidMasterNodeCountError.
func IsInvalidMasterNodeCount(err error) bool {
	return microerror.Cause(err) == invalidMasterNodeCountError
}

var invalidWorkerNodeCountError = &microerror.Error{
	Kind: "invalidWorkerNodeCountError",
}

// IsInvalidWorkerNodeCount asserts invalidWorkerNodeCountError.
func IsInvalidWorkerNodeCount(err error) bool {
	return microerror.Cause(err) == invalidWorkerNodeCountError
}

var executionFailedError = &microerror.Error{
	Kind: "executionFailedError",
}

// IsExecutionFailedError asserts isExecutionFailedError.
func IsExecutionFailedError(err error) bool {
	return microerror.Cause(err) == executionFailedError
}

var notFoundError = &microerror.Error{
	Kind: "notFoundError",
}

// IsNotFound asserts NotFoundError.
func IsNotFound(err error) bool {
	return microerror.Cause(err) == notFoundError
}

var tooManyResultsError = &microerror.Error{
	Kind: "tooManyResultsError",
}

// IsTooManyResults asserts tooManyResultsError.
func IsTooManyResults(err error) bool {
	return microerror.Cause(err) == tooManyResultsError
}

var wrongTypeError = &microerror.Error{
	Kind: "wrongTypeError",
}

// IsWrongType asserts wrongTypeError.
func IsWrongType(err error) bool {
	return microerror.Cause(err) == wrongTypeError
}

// IsStackNotFound asserts stack not found error from upstream's API message
//
// FIXME: The validation error returned by the CloudFormation API doesn't make
// things easy to check, other than looking for the returned string. There's no
// constant in aws go sdk for defining this string, it comes from the service.
func IsStackNotFound(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(microerror.Cause(err).Error(), "does not exist")
}
