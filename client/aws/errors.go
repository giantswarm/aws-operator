package aws

import "github.com/giantswarm/microerror"
import "strings"

const (
	AlreadyAssociated        = "Resource.AlreadyAssociated"
	InvalidSubnetConflict    = "InvalidSubnet.Conflict"
	RoleDuplicate            = "EntityAlreadyExists: Role"
	KeyPairDuplicate         = "InvalidKeyPair.Duplicate"
	SecurityGroupDuplicate   = "InvalidGroup.Duplicate"
	ELBAlreadyExists         = "DuplicateLoadBalancerName"
	ELBConfigurationMismatch = "already exists and it is configured with different parameters"
)

var malformedAmazonAccountIDError = microerror.New("malformed amazon account ID")

// IsMalformedAmazonAccountID asserts malformedAmazonAccountIDError.
func IsMalformedAmazonAccountID(err error) bool {
	return microerror.Cause(err) == malformedAmazonAccountIDError
}

var wrongAmazonAccountIDLengthError = microerror.New("wrong amazon account ID length")

// IsWrongAmazonIDLength asserts wrongAmazonAccountIDLengthError.
func IsWrongAmazonAccountIDLength(err error) bool {
	return microerror.Cause(err) == wrongAmazonAccountIDLengthError
}

var emptyAmazonAccountIDError = microerror.New("empty amazon account ID")

// IsEmptyAmazonAccountID asserts emptyAmazonAccountIDError.
func IsEmptyAmazonAccountID(err error) bool {
	return microerror.Cause(err) == emptyAmazonAccountIDError
}

// IsIAMRoleDuplicateError checks for duplicate IAM Role errors.
func IsIAMRoleDuplicateError(err error) bool {
	return strings.Contains(err.Error(), RoleDuplicate)
}
