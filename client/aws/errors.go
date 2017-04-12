package aws

import "github.com/juju/errgo"

const (
	AlreadyAssociated        = "Resource.AlreadyAssociated"
	InvalidSubnetConflict    = "InvalidSubnet.Conflict"
	BucketAlreadyExists      = "BucketAlreadyExists"
	BucketAlreadyOwnedByYou  = "BucketAlreadyOwnedByYou"
	KeyPairDuplicate         = "InvalidKeyPair.Duplicate"
	SecurityGroupDuplicate   = "InvalidGroup.Duplicate"
	ELBAlreadyExists         = "DuplicateLoadBalancerName"
	ELBConfigurationMismatch = "already exists and it is configured with different parameters"
)

var (
	malformedAmazonAccountIDError   = errgo.New("amazon account ID can only contain numbers")
	wrongAmazonAccountIDLengthError = errgo.New("amazon account ID has the wrong size")
	emptyAmazonAccountIDError       = errgo.New("amazon account ID cannot be empty")
)
