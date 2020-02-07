package cloudformation

import (
	"fmt"
	"strings"

	"github.com/giantswarm/microerror"
)

var invalidConfigError = &microerror.Error{
	Kind: "invalidConfigError",
}

// IsInvalidConfig asserts invalidConfigError.
func IsInvalidConfig(err error) bool {
	return microerror.Cause(err) == invalidConfigError
}

var outputNotFoundError = &microerror.Error{
	Kind: "outputNotFoundError",
}

// IsOutputNotFound asserts outputNotFoundError.
func IsOutputNotFound(err error) bool {
	return microerror.Cause(err) == outputNotFoundError
}

var outputsNotAccessibleError = &microerror.Error{
	Kind: "outputsNotAccessibleError",
}

// IsOutputsNotAccessible asserts outputsNotAccessibleError.
func IsOutputsNotAccessible(err error) bool {
	return microerror.Cause(err) == outputsNotAccessibleError
}

var stackNotFoundError = &microerror.Error{
	Kind: "stackNotFoundError",
}

// IsStackNotFound asserts stackNotFoundError and stack not found errors from
// the upstream's API message.
//
// FIXME: The validation error returned by the CloudFormation API doesn't make
// things easy to check, other than looking for the returned string. There's no
// constant in the AWS golang SDK for defining this string, it comes from the
// service. This is the same in setup/error.go.
func IsStackNotFound(err error) bool {
	if err == nil {
		return false
	}

	c := microerror.Cause(err)

	fmt.Printf("%#v\n", c)
	fmt.Printf("%#v\n", c.Error())

	if strings.Contains(c.Error(), "does not exist") {
		return true
	}

	if c == stackNotFoundError {
		return true
	}

	return false
}

var tooManyStacksError = &microerror.Error{
	Kind: "tooManyStacksError",
}

// IsTooManyStacks asserts tooManyStacksError.
func IsTooManyStacks(err error) bool {
	return microerror.Cause(err) == tooManyStacksError
}
