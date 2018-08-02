package framework

import "github.com/giantswarm/microerror"

var invalidConfigError = &microerror.Error{
	Kind: "invalidConfigError",
}

// IsInvalidConfig asserts invalidConfigError.
func IsInvalidConfig(err error) bool {
	return microerror.Cause(err) == invalidConfigError
}

var missingVaultTokenError = &microerror.Error{
	Kind: "missingVaultTokenError",
}

// IsMissingVaultToken asserts missingVaultTokenError.
func IsMissingVaultToken(err error) bool {
	return microerror.Cause(err) == missingVaultTokenError
}

var notFoundError = &microerror.Error{
	Kind: "notFoundError",
}

// IsNotFound asserts notFoundError.
func IsNotFound(err error) bool {
	return microerror.Cause(err) == notFoundError
}

var tooManyResultsError = &microerror.Error{
	Kind: "tooManyResultsError",
}

// IsTooManyResults asserts invalidConfigError.
func IsTooManyResults(err error) bool {
	return microerror.Cause(err) == tooManyResultsError
}

var unexpectedStatusPhaseError = &microerror.Error{
	Kind: "unexpectedStatusPhaseError",
}

// IsUnexpectedStatusPhase asserts notFoundError.
func IsUnexpectedStatusPhase(err error) bool {
	return microerror.Cause(err) == unexpectedStatusPhaseError
}

var waitError = &microerror.Error{
	Kind: "waitError",
}

// IsWait asserts waitError.
func IsWait(err error) bool {
	return microerror.Cause(err) == waitError
}

var waitTimeoutError = &microerror.Error{
	Kind: "waitTimeoutError",
}

// IsWaitTimeout asserts waitTimeoutError.
func IsWaitTimeout(err error) bool {
	return microerror.Cause(err) == waitTimeoutError
}
