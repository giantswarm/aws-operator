package aws

import "github.com/juju/errgo"

const (
	AlreadyAssociated       = "Resource.AlreadyAssociated"
	BucketAlreadyExists     = "BucketAlreadyExists"
	BucketAlreadyOwnedByYou = "BucketAlreadyOwnedByYou"
	InvalidSubnetConflict   = "InvalidSubnet.Conflict"
	KeyPairDuplicate        = "InvalidKeyPair.Duplicate"
	SecurityGroupDuplicate  = "InvalidGroup.Duplicate"
)

var (
	malformedAmazonAccountIDError   = errgo.New("amazon account ID can only contain numbers")
	wrongAmazonAccountIDLengthError = errgo.New("amazon account ID has the wrong size")
	emptyAmazonAccountIDError       = errgo.New("amazon account ID cannot be empty")
)
