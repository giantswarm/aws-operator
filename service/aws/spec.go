package aws

import (
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/aws/aws-sdk-go/service/sts"
)

type Clients struct {
	KMS KMSClient
	STS STSClient
}

// KMSClient describes the methods required to be implemented by a KMS AWS client.
type KMSClient interface {
	DescribeKey(*kms.DescribeKeyInput) (*kms.DescribeKeyOutput, error)
}

// STSClient describes the methods required to be implemented by a STS AWS client.
type STSClient interface {
	GetCallerIdentity(*sts.GetCallerIdentityInput) (*sts.GetCallerIdentityOutput, error)
}
