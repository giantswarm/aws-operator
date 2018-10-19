package aws

import "strings"

const (
	AlreadyAssociated        = "Resource.AlreadyAssociated"
	InvalidSubnetConflict    = "InvalidSubnet.Conflict"
	RoleDuplicate            = "EntityAlreadyExists: Role"
	KeyPairDuplicate         = "InvalidKeyPair.Duplicate"
	RouteDuplicate           = "RouteAlreadyExists"
	SecurityGroupDuplicate   = "InvalidGroup.Duplicate"
	ELBAlreadyExists         = "DuplicateLoadBalancerName"
	ELBConfigurationMismatch = "already exists and it is configured with different parameters"
)

// IsIAMRoleDuplicateError checks for duplicate IAM Role errors.
func IsIAMRoleDuplicateError(err error) bool {
	return strings.Contains(err.Error(), RoleDuplicate)
}

// IsRouteDuplicateError checks for duplicate Route errors.
func IsRouteDuplicateError(err error) bool {
	return strings.Contains(err.Error(), RouteDuplicate)
}
