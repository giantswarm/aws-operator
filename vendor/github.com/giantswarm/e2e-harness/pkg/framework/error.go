package framework

import "github.com/giantswarm/microerror"

var invalidConfigError = microerror.New("invalid config")

// IsInvalidConfig asserts invalidConfigError.
func IsInvalidConfig(err error) bool {
	return microerror.Cause(err) == invalidConfigError
}

var missingVaultTokenError = microerror.New("missing Vault token")

// IsMissingVaultToken asserts missingVaultTokenError.
func IsMissingVaultToken(err error) bool {
	return microerror.Cause(err) == missingVaultTokenError
}

var notFoundError = microerror.New("not found")

// IsNotFound asserts notFoundError.
func IsNotFound(err error) bool {
	return microerror.Cause(err) == notFoundError
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

var waitError = microerror.New("wait")

// IsWait asserts waitError.
func IsWait(err error) bool {
	return microerror.Cause(err) == waitError
}

var waitTimeoutError = microerror.New("wait timeout")

// IsWaitTimeout asserts waitTimeoutError.
func IsWaitTimeout(err error) bool {
	return microerror.Cause(err) == waitTimeoutError
}
