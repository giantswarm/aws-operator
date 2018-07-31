package collector

import (
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/giantswarm/microerror"
)

const (
	// trustedAdvisorSubscriptionRequiredExceptionCode is the error code returned
	// if Trusted Advisor is not supported (support plan is not Business or Enterprise).
	trustedAdvisorSubscriptionRequiredExceptionCode = "SubscriptionRequiredException"
)

var invalidConfigError = &microerror.Error{
	Kind: "invalidConfigError",
}

// IsInvalidConfig asserts invalidConfigError.
func IsInvalidConfig(err error) bool {
	return microerror.Cause(err) == invalidConfigError
}

var invalidResourceError = &microerror.Error{
	Kind: "invalidResourceError",
}

// IsInvalidResource asserts invalidResourceError.
func IsInvalidResource(err error) bool {
	return microerror.Cause(err) == invalidResourceError
}

var nilLimitError = &microerror.Error{
	Kind: "nilLimitError",
}

// IsNilLimit asserts nilLimitError.
func IsNilLimit(err error) bool {
	return microerror.Cause(err) == nilLimitError
}

var nilUsageError = &microerror.Error{
	Kind: "nilUsageError",
}

// IsNilUsage asserts nilUsageError.
func IsNilUsage(err error) bool {
	return microerror.Cause(err) == nilUsageError
}

// IsUnsupportedPlan asserts that an error is due to Trusted Advisor
// not being available with the current support plan.
func IsUnsupportedPlan(err error) bool {
	c := microerror.Cause(err)
	aerr, ok := c.(awserr.Error)
	if !ok {
		return false
	}

	if aerr.Code() == trustedAdvisorSubscriptionRequiredExceptionCode {
		return true
	}

	return false
}
