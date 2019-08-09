package cleanupsecuritygroups

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws/awserr"
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

var dependencyViolationError = &microerror.Error{
	Kind: "dependencyViolationError",
}

// IsDependencyViolation asserts dependencyViolationError.
func IsDependencyViolation(err error) bool {
	c := microerror.Cause(err)

	if c == dependencyViolationError {
		return true
	}

	aerr, ok := c.(awserr.Error)
	if !ok {
		return false
	}
	fmt.Printf("\n")
	fmt.Printf("\n")
	fmt.Printf("\n")
	fmt.Printf("code: %#v\n", aerr.Code())
	fmt.Printf("erro: %#v\n", aerr.Error())
	fmt.Printf("mess: %#v\n", aerr.Message())
	fmt.Printf("orig: %#v\n", aerr.OrigErr())
	fmt.Printf("\n")
	fmt.Printf("\n")
	fmt.Printf("\n")
	if aerr.Code() == "DependencyViolation" {
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
