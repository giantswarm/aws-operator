package kms

import "github.com/aws/aws-sdk-go/service/kms"

const (
	pendingDeletionWindow = 7
)

type KMSClient interface {
	Encrypt(*kms.EncryptInput) (*kms.EncryptOutput, error)
}
