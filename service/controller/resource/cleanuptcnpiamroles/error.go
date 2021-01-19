package cleanuptcnpiamroles

import (
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

var notFoundError = &microerror.Error{
	Kind: "notFoundError",
}

// IsNotFound asserts notFoundError.
//
//     NoSuchEntity: The role with name gs-cluster-apzh0-role-4z8jm cannot be found.
//
func IsNotFound(err error) bool {
	c := microerror.Cause(err)

	if c == nil {
		return false
	}

	if strings.Contains(c.Error(), "cannot be found") {
		return true
	}

	if c == notFoundError {
		return true
	}

	return false
}
