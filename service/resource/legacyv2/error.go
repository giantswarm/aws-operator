package legacyv2

import (
	"strings"

	"github.com/giantswarm/microerror"
)

const (
	// Format for masked idleTimeoutSecondsOutOfRangeError.
	idleTimeoutSecondsOutOfRangeErrorFormat string = "ELB idle timeout seconds %s cannot exceed AWS maximum of 3600"
)

var idleTimeoutSecondsOutOfRangeError = microerror.New("idle timeout seconds out of range")

// IsIdleTimeoutSecondsOutOfRangeError asserts idleTimeoutSecondsOutOfRangeError.
func IsIdleTimeoutSecondsOutOfRangeError(err error) bool {
	return microerror.Cause(err) == idleTimeoutSecondsOutOfRangeError
}

var invalidConfigError = microerror.New("invalid config")

// IsInvalidConfig asserts invalidConfigError.
func IsInvalidConfig(err error) bool {
	return microerror.Cause(err) == invalidConfigError
}

var invalidCloudconfigExtensionNameError = microerror.New("invalid cloudconfig extension name")

//  asserts invalidCloudconfigExtensionNameError.
func IsInvalidCloudconfigExtensionName(err error) bool {
	return microerror.Cause(err) == invalidCloudconfigExtensionNameError
}

var malformedCloudConfigKeyError = microerror.New("malformed key in the cloudconfig")

// IsMalformedCloudConfigKey asserts malformedCloudConfigKeyError.
func IsMalformedCloudConfigKey(err error) bool {
	return microerror.Cause(err) == malformedCloudConfigKeyError
}

var missingCloudConfigKeyError = microerror.New("missing key in the cloudconfig")

// IsMissingCloudConfigKey asserts missingCloudConfigKeyError.
func IsMissingCloudConfigKey(err error) bool {
	return microerror.Cause(err) == missingCloudConfigKeyError
}

var secretsRetrievalFailedError = microerror.New("secrets retrieval failed")

// IsSecretsRetrievalFailed asserts secretsRetrievalFailedError.
func IsSecretsRetrievalFailed(err error) bool {
	return microerror.Cause(err) == secretsRetrievalFailedError
}

// Validation errors

var invalidAvailabilityZoneError = microerror.New("invalid availability zone")

// IsInvalidAvailabilityZone asserts invalidAvailabilityZoneError.
func IsInvalidAvailabilityZone(err error) bool {
	return microerror.Cause(err) == invalidAvailabilityZoneError
}

var workersListEmptyError = microerror.New("workers list empty")

// IsWorkersListEmpty asserts workersListEmptyError.
func IsWorkersListEmpty(err error) bool {
	return microerror.Cause(err) == workersListEmptyError
}

var differentImageIDsError = microerror.New("different image IDs")

// IsDifferentImageIDs assert differentImageIDsError.
func IsDifferentImageIDs(err error) bool {
	return microerror.Cause(err) == differentImageIDsError
}

var differentInstanceTypesError = microerror.New("different instance types")

// IsDifferentInstanceTypes asserts differentInstanceTypesError.
func IsDifferentInstanceTypes(err error) bool {
	return microerror.Cause(err) == differentInstanceTypesError
}

var invalidMasterNodeCountError = microerror.New("invalid master node count")

// IsInvalidMasterNodeCount asserts invalidMasterNodeCountError.
func IsInvalidMasterNodeCount(err error) bool {
	return microerror.Cause(err) == invalidMasterNodeCountError
}

var invalidWorkerNodeCountError = microerror.New("invalid worker node count")

// IsInvalidWorkerNodeCount asserts invalidWorkerNodeCountError.
func IsInvalidWorkerNodeCount(err error) bool {
	return microerror.Cause(err) == invalidWorkerNodeCountError
}

var executionFailedError = microerror.New("execution failed")

// IsExecutionFailedError asserts isExecutionFailedError.
func IsExecutionFailedError(err error) bool {
	return microerror.Cause(err) == executionFailedError
}

var notFoundError = microerror.New("not found")

// IsNotFound asserts NotFoundError.
func IsNotFound(err error) bool {
	return microerror.Cause(err) == notFoundError
}

var tooManyResultsError = microerror.New("too many results")

// IsTooManyResults asserts tooManyResultsError.
func IsTooManyResults(err error) bool {
	return microerror.Cause(err) == tooManyResultsError
}

var wrongTypeError = microerror.New("wrong type")

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
