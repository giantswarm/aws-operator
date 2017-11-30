// +build k8srequired

package integration

import "github.com/giantswarm/microerror"

var waitTimeoutError = microerror.New("waitTimeout")

// IsWaitTimeout asserts invalidConfigError.
func IsWaitTimeout(err error) bool {
	return microerror.Cause(err) == waitTimeoutError
}

var tooManyResultsError = microerror.New("too many results")

// IsTooManyResults asserts invalidConfigError.
func IsTooManyResults(err error) bool {
	return microerror.Cause(err) == tooManyResultsError
}

var unexpectedStatusPhaseError = microerror.New("unexpected status phase")

// IsUnexpectedStatusPhase asserts notFoundError.
func IsUnexpectedStatusPhase(err error) bool {
	return microerror.Cause(err) == unexpectedStatusPhaseError
}

var notFoundError = microerror.New("not found")

// IsNotFound asserts notFoundError.
func IsNotFound(err error) bool {
	return microerror.Cause(err) == notFoundError
}
