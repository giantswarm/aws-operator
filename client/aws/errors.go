package aws

import "github.com/juju/errgo"

const (
	AlreadyAssociated        = "Resource.AlreadyAssociated"
	InvalidSubnetConflict    = "InvalidSubnet.Conflict"
	KeyPairDuplicate         = "InvalidKeyPair.Duplicate"
	SecurityGroupDuplicate   = "InvalidGroup.Duplicate"
	ELBAlreadyExists         = "DuplicateLoadBalancerName"
	ELBConfigurationMismatch = "already exists and it is configured with different parameters"
)

var malformedAmazonAccountIDError = errgo.New("malformed amazon account ID")

// IsMalformedAmazonAccountID asserts malformedAmazonAccountIDError.
func IsMalformedAmazonAccountID(err error) bool {
	return errgo.Cause(err) == malformedAmazonAccountIDError
}

var wrongAmazonAccountIDLengthError = errgo.New("wrong amazon account ID length")

// IsWrongAmazonIDLength asserts wrongAmazonAccountIDLengthError.
func IsWrongAmazonAccountIDLength(err error) bool {
	return errgo.Cause(err) == wrongAmazonAccountIDLengthError
}

var emptyAmazonAccountIDError = errgo.New("empty amazon account ID")

// IsEmptyAmazonAccountID asserts emptyAmazonAccountIDError.
func IsEmptyAmazonAccountID(err error) bool {
	return errgo.Cause(err) == emptyAmazonAccountIDError
}
