package aws

import "strings"

const (
	AlreadyAssociated      = "Resource.AlreadyAssociated"
	RoleDuplicate          = "EntityAlreadyExists: Role"
	KeyPairDuplicate       = "InvalidKeyPair.Duplicate"
	RouteDuplicate         = "RouteAlreadyExists"
	SecurityGroupDuplicate = "InvalidGroup.Duplicate"
)

// IsIAMRoleDuplicateError checks for duplicate IAM Role errors.
func IsIAMRoleDuplicateError(err error) bool {
	return strings.Contains(err.Error(), RoleDuplicate)
}

// IsRouteDuplicateError checks for duplicate Route errors.
func IsRouteDuplicateError(err error) bool {
	return strings.Contains(err.Error(), RouteDuplicate)
}
