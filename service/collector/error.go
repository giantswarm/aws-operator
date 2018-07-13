package collector

import (
	"strings"

	"github.com/giantswarm/microerror"
)

const (
	// trustedAdvisorUnsupportedErrorMessage is the error message returned
	// if Trusted Advisor is not supported (support plan is not Business or Enterprise).
	trustedAdvisorUnsupportedErrorMessage = "AWS Premium Support Subscription is required to use this service."
)

var invalidConfigError = microerror.New("invalid config")

// IsInvalidConfig asserts invalidConfigError.
func IsInvalidConfig(err error) bool {
	return microerror.Cause(err) == invalidConfigError
}

var invalidResourceError = microerror.New("invalid resource")

// IsInvalidResource asserts invalidResourceError.
func IsInvalidResource(err error) bool {
	return microerror.Cause(err) == invalidResourceError
}

var nilLimitError = microerror.New("nil limit")

// IsNilLimit asserts nilLimitError.
func IsNilLimit(err error) bool {
	return microerror.Cause(err) == nilLimitError
}

var nilUsageError = microerror.New("nil usage")

// IsNilUsage asserts nilUsageError.
func IsNilUsage(err error) bool {
	return microerror.Cause(err) == nilUsageError
}

// IsUnsupportedPlan asserts that an error is due to Trusted Advisor
// not being available with the current support plan.
func IsUnsupportedPlan(err error) bool {
	return strings.Contains(err.Error(), trustedAdvisorUnsupportedErrorMessage)
}
