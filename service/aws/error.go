package aws

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

var wrongAmazonAccountIDLengthError = &microerror.Error{
	Kind: "wrongAmazonAccountIDLengthError",
}

// IsWrongAmazonIDLength asserts wrongAmazonAccountIDLengthError.
func IsWrongAmazonAccountIDLength(err error) bool {
	return microerror.Cause(err) == wrongAmazonAccountIDLengthError
}
