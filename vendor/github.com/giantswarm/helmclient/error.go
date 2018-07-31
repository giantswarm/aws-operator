package helmclient

import (
	"strings"

	"github.com/giantswarm/microerror"
)

const (
	cannotReuseReleaseErrorPrefix = "cannot re-use"
)

var cannotReuseReleaseError = &microerror.Error{
	Kind: "cannotReuseReleaseError",
}

// IsCannotReuseRelease asserts cannotReuseReleaseError.
func IsCannotReuseRelease(err error) bool {
	if err == nil {
		return false
	}

	c := microerror.Cause(err)

	if strings.Contains(c.Error(), cannotReuseReleaseErrorPrefix) {
		return true
	}
	if c == cannotReuseReleaseError {
		return true
	}

	return false
}

var executionFailedError = &microerror.Error{
	Kind: "executionFailedError",
}

// IsExecutionFailed asserts executionFailedError.
func IsExecutionFailed(err error) bool {
	return microerror.Cause(err) == executionFailedError
}

var invalidConfigError = &microerror.Error{
	Kind: "invalidConfigError",
}

// IsInvalidConfig asserts invalidConfigError.
func IsInvalidConfig(err error) bool {
	return microerror.Cause(err) == invalidConfigError
}

const (
	invalidGZipHeaderErrorPrefix = "gzip: invalid header"
)

var invalidGZipHeaderError = &microerror.Error{
	Kind: "invalidGZipHeaderError",
}

// IsInvalidGZipHeader asserts invalidGZipHeaderError.
func IsInvalidGZipHeader(err error) bool {
	if err == nil {
		return false
	}

	c := microerror.Cause(err)

	if strings.HasPrefix(c.Error(), invalidGZipHeaderErrorPrefix) {
		return true
	}
	if c == invalidGZipHeaderError {
		return true
	}

	return false
}

var podNotFoundError = &microerror.Error{
	Kind: "podNotFoundError",
}

// IsPodNotFound asserts podNotFoundError.
func IsPodNotFound(err error) bool {
	return microerror.Cause(err) == podNotFoundError
}

const (
	releaseNotFoundErrorPrefix = "No such release:"
	releaseNotFoundErrorSuffix = "not found"
)

var releaseNotFoundError = &microerror.Error{
	Kind: "releaseNotFoundError",
}

// IsReleaseNotFound asserts releaseNotFoundError.
func IsReleaseNotFound(err error) bool {
	if err == nil {
		return false
	}

	c := microerror.Cause(err)

	if strings.HasPrefix(c.Error(), releaseNotFoundErrorPrefix) {
		return true
	}
	if strings.HasSuffix(c.Error(), releaseNotFoundErrorSuffix) {
		return true
	}
	if c == releaseNotFoundError {
		return true
	}

	return false
}

var testReleaseFailureError = &microerror.Error{
	Kind: "testReleaseFailureError",
}

// IsTestReleaseFailure asserts testReleaseFailureError.
func IsTestReleaseFailure(err error) bool {
	return microerror.Cause(err) == testReleaseFailureError
}

var testReleaseTimeoutError = &microerror.Error{
	Kind: "testReleaseTimeoutError",
}

// IsTestReleaseTimeout asserts testReleaseTimeoutError.
func IsTestReleaseTimeout(err error) bool {
	return microerror.Cause(err) == testReleaseTimeoutError
}

var tillerInstallationFailedError = &microerror.Error{
	Kind: "tillerInstallationFailedError",
}

// IsTillerInstallationFailed asserts tillerInstallationFailedError.
func IsTillerInstallationFailed(err error) bool {
	return microerror.Cause(err) == tillerInstallationFailedError
}

var tooManyResultsError = &microerror.Error{
	Kind: "tooManyResultsError",
}

// IsTooManyResults asserts tooManyResultsError.
func IsTooManyResults(err error) bool {
	return microerror.Cause(err) == tooManyResultsError
}
