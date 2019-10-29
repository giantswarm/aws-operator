package credential

import "github.com/giantswarm/microerror"

var arnNotFound = &microerror.Error{
	Kind: "arnNotFound",
}

// IsArnNotFoundError asserts arnNotFound.
func IsArnNotFoundError(err error) bool {
	return microerror.Cause(err) == arnNotFound
}

var credentialNameEmpty = &microerror.Error{
	Kind: "credentialNameEmpty",
}

// IsArnNotFoundError asserts credentialNameEmpty.
func IsCredentialNameEmptyError(err error) bool {
	return microerror.Cause(err) == credentialNameEmpty
}

var credentialNamespaceEmpty = &microerror.Error{
	Kind: "credentialNamespaceEmpty",
}

// IsArnNotFoundError asserts credentialNamespaceEmpty.
func IsCredentialNamespaceEmptyError(err error) bool {
	return microerror.Cause(err) == credentialNamespaceEmpty
}
