package aws

import "github.com/juju/errgo"

const (
	BucketAlreadyExists     = "BucketAlreadyExists"
	BucketAlreadyOwnedByYou = "BucketAlreadyOwnedByYou"
	KeyPairDuplicate        = "InvalidKeyPair.Duplicate"
	SecurityGroupDuplicate  = "InvalidGroup.Duplicate"
)

var (
	malformedAmazonAccountIDError   = errgo.New("amazon account ID can only contain numbers")
	wrongAmazonAccountIDLengthError = errgo.New("amazon account ID has the wrong size")
	emptyAmazonAccountIDError       = errgo.New("amazon account ID cannot be empty")
)
