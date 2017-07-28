package kubeadmtokentpr

import (
	"github.com/juju/errgo"
)

var tokenRetrievalFailedError = errgo.New("token retrieval failed")

// IsTokenRetrievalFailed asserts tokenRetrievalFailedError.
func IsTokenRetrievalFailed(err error) bool {
	return errgo.Cause(err) == tokenRetrievalFailedError
}
