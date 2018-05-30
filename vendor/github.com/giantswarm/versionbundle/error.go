package versionbundle

import (
	"github.com/giantswarm/microerror"
)

var bundleNotFoundError = microerror.New("bundle not found")

// IsBundleNotFound asserts bundleNotFoundError.
func IsBundleNotFound(err error) bool {
	return microerror.Cause(err) == bundleNotFoundError
}

var executionFailedError = microerror.New("execution failed")

// IsExecutionFailed asserts executionFailedError.
func IsExecutionFailed(err error) bool {
	return microerror.Cause(err) == executionFailedError
}

var invalidBundleError = microerror.New("invalid bundle")

// IsInvalidBundleError asserts invalidBundleError.
func IsInvalidBundleError(err error) bool {
	return microerror.Cause(err) == invalidBundleError
}

var invalidBundlesError = microerror.New("invalid bundles")

// IsInvalidBundlesError asserts invalidBundlesError.
func IsInvalidBundlesError(err error) bool {
	return microerror.Cause(err) == invalidBundlesError
}

var invalidChangelogError = microerror.New("invalid changelog")

// IsInvalidChangelog asserts invalidChangelogError.
func IsInvalidChangelog(err error) bool {
	return microerror.Cause(err) == invalidChangelogError
}

var invalidComponentError = microerror.New("invalid component")

// IsInvalidComponent asserts invalidComponentError.
func IsInvalidComponent(err error) bool {
	return microerror.Cause(err) == invalidComponentError
}

var invalidConfigError = microerror.New("invalid config")

// IsInvalidConfig asserts invalidConfigError.
func IsInvalidConfig(err error) bool {
	return microerror.Cause(err) == invalidConfigError
}

var invalidReleaseError = microerror.New("invalid release")

// IsInvalidRelease asserts invalidReleaseError.
func IsInvalidRelease(err error) bool {
	return microerror.Cause(err) == invalidReleaseError
}
