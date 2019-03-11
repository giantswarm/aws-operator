package peerrolearn

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/giantswarm/microerror"
)

var invalidConfigError = &microerror.Error{
	Kind: "invalidConfigError",
}

// IsInvalidConfig asserts invalidConfigError.
func IsInvalidConfig(err error) bool {
	return microerror.Cause(err) == invalidConfigError
}

var notFoundError = &microerror.Error{
	Kind: "notFoundError",
}

// IsNotFound asserts notFoundError.
func IsNotFound(err error) bool {
	c := microerror.Cause(err)

	aerr, ok := c.(awserr.Error)
	fmt.Printf("%#v\n", aerr)
	if ok {
		fmt.Printf("%#v\n", aerr.Code())
		fmt.Printf("%#v\n", aerr.Error())
		fmt.Printf("%#v\n", aerr.Message())
		fmt.Printf("%#v\n", aerr.OrigErr())
		if aerr.Code() == "NoSuchEntity" {
			return true
		}
	}

	if c == notFoundError {
		return true
	}

	return false
}
