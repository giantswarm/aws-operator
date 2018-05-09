package credential

import "github.com/giantswarm/microerror"

var malformedRole = microerror.New("malformed role")

// IsMalformedRoleError asserts malformedRole.
func IsMalformedRoleError(err error) bool {
	return microerror.Cause(err) == malformedRole
}

var roleNotFound = microerror.New("role not found")

// IsRoleNotFoundError asserts roleNotFound.
func IsRoleNotFoundError(err error) bool {
	return microerror.Cause(err) == roleNotFound
}
