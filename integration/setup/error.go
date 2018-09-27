package setup

import "github.com/giantswarm/microerror"

var notFoundError = &microerror.Error{
	Kind: "notFoundError",
}

// IsNotFound asserts notFoundError.
func IsNotFound(err error) bool {
	return microerror.Cause(err) == notFoundError
}

var stillExistsError = &microerror.Error{
	Kind: "stillExistsError",
}

// IsStillExists asserts stillExistsError.
func IsStillExists(err error) bool {
	return microerror.Cause(err) == stillExistsError
}

var tooManyResultsError = &microerror.Error{
	Kind: "tooManyResultsError",
}

// IsTooManyResults asserts tooManyResultsError.
func IsTooManyResults(err error) bool {
	return microerror.Cause(err) == tooManyResultsError
}
