package tccpnlifecycle

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

// IsInsserts invalidConfigError.
func IsInvalidConfig(err error) bool {
	return microerror.Cause(err) == invalidConfigError
}

var invalidProviderIDError = &microerror.Error{
	Kind: "invalidProviderID",
}

// IsInvalidProviderID asserts invalidConfigError.
func IsInvalidProviderID(err error) bool {
	return microerror.Cause(err) == invalidProviderIDError
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
