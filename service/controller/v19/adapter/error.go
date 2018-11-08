package adapter

import "github.com/giantswarm/microerror"

var emptyAmazonAccountIDError = &microerror.Error{
	Kind: "emptyAmazonAccountIDError",
}

// IsEmptyAmazonAccountID asserts emptyAmazonAccountIDError.
func IsEmptyAmazonAccountID(err error) bool {
	return microerror.Cause(err) == emptyAmazonAccountIDError
}

var invalidConfigError = &microerror.Error{
	Kind: "invalidConfigError",
}

// IsInvalidConfig asserts invalidConfigError.
func IsInvalidConfig(err error) bool {
	return microerror.Cause(err) == invalidConfigError
}

var malformedAmazonAccountIDError = &microerror.Error{
	Kind: "malformedAmazonAccountIDError",
}

// IsMalformedAmazonAccountID asserts malformedAmazonAccountIDError.
func IsMalformedAmazonAccountID(err error) bool {
	return microerror.Cause(err) == malformedAmazonAccountIDError
}

var notFoundError = &microerror.Error{
	Kind: "notFoundError",
}

// IsNotFound asserts notFoundError
func IsNotFound(err error) bool {
	return microerror.Cause(err) == notFoundError
}

var wrongAmazonAccountIDLengthError = &microerror.Error{
	Kind: "wrongAmazonAccountIDLengthError",
}

// IsWrongAmazonAccountIDLength asserts wrongAmazonAccountIDLengthError.
func IsWrongAmazonAccountIDLength(err error) bool {
	return microerror.Cause(err) == wrongAmazonAccountIDLengthError
}

var wrongTypeError = &microerror.Error{
	Kind: "wrongTypeError",
}

// IsWrongType asserts wrongTypeError.
func IsWrongType(err error) bool {
	return microerror.Cause(err) == wrongTypeError
}

var tooFewResultsError = &microerror.Error{
	Kind: "tooFewResultsError",
}

// IsTooFewResults asserts tooFewResultsError.
func IsTooFewResults(err error) bool {
	return microerror.Cause(err) == tooFewResultsError
}

var tooManyResultsError = &microerror.Error{
	Kind: "tooManyResultsError",
}

// IsTooManyResults asserts tooManyResultsError.
func IsTooManyResults(err error) bool {
	return microerror.Cause(err) == tooManyResultsError
}
