package statusresource

import (
	"github.com/giantswarm/microerror"
	"github.com/prometheus/client_golang/prometheus"
)

var alreadyRegisteredError = &microerror.Error{
	Kind: "alreadyRegisteredError",
}

// IsAlreadyRegisteredError asserts alreadyRegisteredError.
func IsAlreadyRegisteredError(err error) bool {
	c := microerror.Cause(err)
	_, ok := c.(prometheus.AlreadyRegisteredError)
	if ok {
		return true
	}
	if c == alreadyRegisteredError {
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

var missingLabelError = &microerror.Error{
	Kind: "missingLabelError",
}

// IsMissingLabel asserts missingLabelError.
func IsMissingLabel(err error) bool {
	return microerror.Cause(err) == missingLabelError
}

var notFoundError = &microerror.Error{
	Kind: "notFoundError",
}

// IsNotFound asserts notFoundError.
func IsNotFound(err error) bool {
	return microerror.Cause(err) == notFoundError
}
