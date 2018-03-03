package cloudformation

import (
	"strings"

	"github.com/giantswarm/microerror"
)

var invalidConfigError = microerror.New("invalid config")

// IsInvalidConfig asserts invalidConfigError.
func IsInvalidConfig(err error) bool {
	return microerror.Cause(err) == invalidConfigError
}

var outputNotFoundError = microerror.New("output not found")

// IsOutputNotFound asserts outputNotFoundError.
func IsOutputNotFound(err error) bool {
	return microerror.Cause(err) == outputNotFoundError
}

var outputsNotAccessibleError = microerror.New("outputs not accessible")

// IsOutputsNotAccessible asserts outputsNotAccessibleError.
func IsOutputsNotAccessible(err error) bool {
	return microerror.Cause(err) == outputsNotAccessibleError
}

var stackNotFoundError = microerror.New("stack not found")

// IsStackNotFound asserts stackNotFoundError and stack not found errors from
// the upstream's API message.
//
// FIXME: The validation error returned by the CloudFormation API doesn't make
// things easy to check, other than looking for the returned string. There's no
// constant in the AWS golang SDK for defining this string, it comes from the
// service.
func IsStackNotFound(err error) bool {
	if err == nil {
		return false
	}

	if strings.Contains(microerror.Cause(err).Error(), "does not exist") {
		return true
	}

	if microerror.Cause(err) == stackNotFoundError {
		return true
	}

	return false
}

var tooManyStacksError = microerror.New("too many stacks")

// IsTooManyStacks asserts tooManyStacksError.
func IsTooManyStacks(err error) bool {
	return microerror.Cause(err) == tooManyStacksError
}
