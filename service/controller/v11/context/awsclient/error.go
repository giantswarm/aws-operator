package awsclient

import "github.com/giantswarm/microerror"

var clientsNotFound = microerror.New("clients not found")

// IsClientsNotFoundError asserts clientsNotFound.
func IsClientsNotFoundError(err error) bool {
	return microerror.Cause(err) == clientsNotFound
}
