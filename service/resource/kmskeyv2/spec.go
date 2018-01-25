package kmskeyv2

import "github.com/aws/aws-sdk-go/service/kms"

const (
	pendingDeletionWindow = 7
)

type KMSKeyState struct {
	KeyID    string
	KeyAlias string
}

type Clients struct {
	KMS KMSClient
}

type KMSClient interface {
	CreateKey(*kms.CreateKeyInput) (*kms.CreateKeyOutput, error)
	CreateAlias(*kms.CreateAliasInput) (*kms.CreateAliasOutput, error)
	DescribeKey(*kms.DescribeKeyInput) (*kms.DescribeKeyOutput, error)
	DeleteAlias(*kms.DeleteAliasInput) (*kms.DeleteAliasOutput, error)
	EnableKeyRotation(*kms.EnableKeyRotationInput) (*kms.EnableKeyRotationOutput, error)
	ScheduleKeyDeletion(*kms.ScheduleKeyDeletionInput) (*kms.ScheduleKeyDeletionOutput, error)
	TagResource(*kms.TagResourceInput) (*kms.TagResourceOutput, error)
}
