package release

import (
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
func IsNotFound(err error) bool {
	return microerror.Cause(err) == notFoundError
}

var releaseAlreadyExistsError = &microerror.Error{
	Kind: "releaseAlreadyExistsError",
}

// IsReleaseAlreadyExists asserts releaseAlreadyExistsError.
func IsReleaseAlreadyExists(err error) bool {
	return microerror.Cause(err) == releaseAlreadyExistsError
}

var releaseNotFoundError = &microerror.Error{
	Kind: "releaseNotFoundError",
}

// IsReleaseNotFound asserts releaseNotFoundError.
func IsReleaseNotFound(err error) bool {
	return microerror.Cause(err) == releaseNotFoundError
}

var releaseStatusNotMatchingError = &microerror.Error{
	Kind: "releaseStatusNotMatchingError",
}

// IsReleaseStatusNotMatching asserts releaseStatusNotMatchingError
func IsReleaseStatusNotMatching(err error) bool {
	return microerror.Cause(err) == releaseStatusNotMatchingError
}

var releaseVersionNotMatchingError = &microerror.Error{
	Kind: "releaseVersionNotMatchingError",
}

// IsReleaseVersionNotMatching asserts releaseVersionNotMatchingError
func IsReleaseVersionNotMatching(err error) bool {
	return microerror.Cause(err) == releaseVersionNotMatchingError
}

var tarballNotFoundError = &microerror.Error{
	Kind: "tarballNotFoundError",
}

// IsTarballNotFound asserts tarballNotFoundError.
func IsTarballNotFound(err error) bool {
	return microerror.Cause(err) == tarballNotFoundError
}

var tillerNotFoundError = &microerror.Error{
	Kind: "tillerNotFoundError",
}

// IsTillerNotFound asserts tillerNotFoundError.
func IsTillerNotFound(err error) bool {
	return microerror.Cause(err) == tillerNotFoundError
}

var waitError = &microerror.Error{
	Kind: "waitError",
}

// IsWait asserts waitError.
func IsWait(err error) bool {
	return microerror.Cause(err) == waitError
}
