package helmclient

import (
	"strings"

	"github.com/giantswarm/microerror"
)

const (
	releaseNotFoundErrorPrefix = "No such release:"
	releaseNotFoundErrorSuffix = "not found"
)

var invalidConfigError = microerror.New("invalid config")

// IsInvalidConfig asserts invalidConfigError.
func IsInvalidConfig(err error) bool {
	return microerror.Cause(err) == invalidConfigError
}

var notFoundError = microerror.New("not found")

// IsNotFound asserts notFoundError.
func IsNotFound(err error) bool {
	return microerror.Cause(err) == notFoundError
}

var releaseNotFoundError = microerror.New("release not found")

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

var tooManyResultsError = microerror.New("too many results")

// IsTooManyResults asserts tooManyResultsError.
func IsTooManyResults(err error) bool {
	return microerror.Cause(err) == tooManyResultsError
}
