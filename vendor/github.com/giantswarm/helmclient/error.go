package helmclient

import (
	"regexp"
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

var notFoundError = &microerror.Error{
	Kind: "notFoundError",
}

// IsNotFound asserts notFoundError.
func IsNotFound(err error) bool {
	return microerror.Cause(err) == notFoundError
}

var (
	releaseAlreadyExistsRegexp = regexp.MustCompile(`release named \S+ already exists`)
)

var releaseAlreadyExistsError = &microerror.Error{
	Kind: "releaseAlreadyExistsError",
}

// IsReleaseAlreadyExists asserts releaseAlreadyExistsError.
func IsReleaseAlreadyExists(err error) bool {
	if err == nil {
		return false
	}

	c := microerror.Cause(err)

	if c == releaseAlreadyExistsError {
		return true
	}
	if releaseAlreadyExistsRegexp.MatchString(c.Error()) {
		return true
	}

	return false
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

var (
	tarballNotFoundRegexp = regexp.MustCompile(`stat \S+: no such file or directory`)
)

var tarballNotFoundError = &microerror.Error{
	Kind: "tarballNotFoundError",
}

// IsTarballNotFound asserts tarballNotFoundError.
func IsTarballNotFound(err error) bool {
	if err == nil {
		return false
	}

	c := microerror.Cause(err)

	if c == tarballNotFoundError {
		return true
	}
	if tarballNotFoundRegexp.MatchString(c.Error()) {
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

var tillerNotFoundError = &microerror.Error{
	Kind: "tillerNotFoundError",
}

// IsTillerNotFound asserts tillerNotFoundError.
func IsTillerNotFound(err error) bool {
	return microerror.Cause(err) == tillerNotFoundError
}

var tillerInvalidVersionError = &microerror.Error{
	Kind: "tillerInvalidVersionError",
}

// IsTillerInvalidVersion asserts tillerInvalidVersionError.
func IsTillerInvalidVersion(err error) bool {
	return microerror.Cause(err) == tillerInvalidVersionError
}

var tooManyResultsError = &microerror.Error{
	Kind: "tooManyResultsError",
}

// IsTooManyResults asserts tooManyResultsError.
func IsTooManyResults(err error) bool {
	return microerror.Cause(err) == tooManyResultsError
}

var (
	yamlConversionFailedErrorText = "error converting YAML to JSON:"
)

var yamlConversionFailedError = &microerror.Error{
	Kind: "yamlConversionFailedError",
}

// IsYamlConversionFailed asserts yamlConversionFailedError.
func IsYamlConversionFailed(err error) bool {
	if err == nil {
		return false
	}

	c := microerror.Cause(err)

	if c == yamlConversionFailedError {
		return true
	}
	if strings.Contains(c.Error(), yamlConversionFailedErrorText) {
		return true
	}

	return false
}
