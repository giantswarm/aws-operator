package create

import (
	"github.com/juju/errgo"
)

var invalidConfigError = errgo.New("invalid config")

// IsInvalidConfig asserts invalidConfigError.
func IsInvalidConfig(err error) bool {
	return errgo.Cause(err) == invalidConfigError
}

var invalidCloudconfigExtensionNameError = errgo.New("invalid cloudconfig extension name")

//  asserts invalidCloudconfigExtensionNameError.
func IsInvalidCloudconfigExtensionName(err error) bool {
	return errgo.Cause(err) == invalidCloudconfigExtensionNameError
}

var malformedCloudConfigKeyError = errgo.New("malformed key in the cloudconfig")

// IsMalformedCloudConfigKey asserts malformedCloudConfigKeyError.
func IsMalformedCloudConfigKey(err error) bool {
	return errgo.Cause(err) == malformedCloudConfigKeyError
}

var missingCloudConfigKeyError = errgo.New("missing key in the cloudconfig")

// IsMissingCloudConfigKey asserts missingCloudConfigKeyError.
func IsMissingCloudConfigKey(err error) bool {
	return errgo.Cause(err) == missingCloudConfigKeyError
}

var secretsRetrievalFailedError = errgo.New("secrets retrieval failed")

// IsSecretsRetrievalFailed asserts secretsRetrievalFailedError.
func IsSecretsRetrievalFailed(err error) bool {
	return errgo.Cause(err) == secretsRetrievalFailedError
}

// Validation errors

var workersListEmptyError = errgo.New("workers list empty")

// IsWorkersListEmpty asserts workersListEmptyError.
func IsWorkersListEmpty(err error) bool {
	return errgo.Cause(err) == workersListEmptyError
}

var differentImageIDsError = errgo.New("different image IDs")

// IsDifferentImageIDs assert differentImageIDsError.
func IsDifferentImageIDs(err error) bool {
	return errgo.Cause(err) == differentImageIDsError
}

var differentInstanceTypesError = errgo.New("different instance types")

// IsDifferentInstanceTypes asserts differentInstanceTypesError.
func IsDifferentInstanceTypes(err error) bool {
	return errgo.Cause(err) == differentInstanceTypesError
}
