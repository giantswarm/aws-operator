package drainerfinalizer

import (
	"regexp"

	"github.com/giantswarm/microerror"
)

var (
	// noActiveLifeCycleActionRegExp is a fuzzy regular expression to match
	// Autoscaling errors which we have to string match due to the lack of proper
	// error types in the AWS SDK.
	noActiveLifeCycleActionRegExp = regexp.MustCompile(`(?im)no.*life.*cycle.*found`)
)

var invalidConfigError = &microerror.Error{
	Kind: "invalidConfigError",
}

// IsInvalidConfig asserts invalidConfigError.
func IsInvalidConfig(err error) bool {
	return microerror.Cause(err) == invalidConfigError
}

var missingAnnotationError = &microerror.Error{
	Kind: "missingAnnotationError",
}

func IsMissingAnnotationError(err error) bool {
	return microerror.Cause(err) == missingAnnotationError
}

var noActiveLifeCycleActionError = &microerror.Error{
	Kind: "noActiveLifeCycleActionError",
}

// IsNoActiveLifeCycleAction asserts noActiveLifeCycleActionError. It also
// checks for some string matching in the error message to figure if the AWS API
// gives the error we expect.
func IsNoActiveLifeCycleAction(err error) bool {
	c := microerror.Cause(err)

	if c == nil {
		return false
	}

	if noActiveLifeCycleActionRegExp.MatchString(c.Error()) {
		return true
	}

	if c == noActiveLifeCycleActionError {
		return true
	}

	return false
}
