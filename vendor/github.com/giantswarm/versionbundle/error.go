package versionbundle

import (
	"github.com/giantswarm/microerror"
)

var invalidAggregatedBundlesError = microerror.New("invalid aggregated bundles")

// IsInvalidAggregatedBundlesError asserts invalidAggregatedBundlesError.
func IsInvalidAggregatedBundlesError(err error) bool {
	return microerror.Cause(err) == invalidAggregatedBundlesError
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

var invalidCapabilityError = microerror.New("invalid capability")

// IsInvalidCapability asserts invalidCapabilityError.
func IsInvalidCapability(err error) bool {
	return microerror.Cause(err) == invalidCapabilityError
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

var invalidDependencyError = microerror.New("invalid dependency")

// IsInvalidDependency asserts invalidDependencyError.
func IsInvalidDependency(err error) bool {
	return microerror.Cause(err) == invalidDependencyError
}
