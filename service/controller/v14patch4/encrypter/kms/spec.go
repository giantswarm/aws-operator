package kms

import "github.com/aws/aws-sdk-go/service/kms"

const (
	// pendingDeletionWindow is the number of days to keep a key on pending
	// deletion state.
	pendingDeletionWindow = 7
)

type KMSClient interface {
	Encrypt(*kms.EncryptInput) (*kms.EncryptOutput, error)
}
