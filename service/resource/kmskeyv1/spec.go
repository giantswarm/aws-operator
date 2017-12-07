package kmskeyv1

import "github.com/aws/aws-sdk-go/service/kms"

type KMSKeyState struct {
	KeyID string
	ARN   string
}

type Clients struct {
	KMS KMSClient
}

type KMSClient interface {
	CreateKey(*kms.CreateKeyInput) (*kms.CreateKeyOutput, error)
	CreateAlias(*kms.CreateAliasInput) (*kms.CreateAliasOutput, error)
	DeleteAlias(*kms.DeleteAliasInput) (*kms.DeleteAliasOutput, error)
	EnableKeyRotation(*kms.EnableKeyRotationInput) (*kms.EnableKeyRotationOutput, error)
	DescribeKey(*kms.DescribeKeyInput) (*kms.DescribeKeyOutput, error)
}
