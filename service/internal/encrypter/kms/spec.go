package kms

import "github.com/aws/aws-sdk-go/service/kms"

type KMSClient interface {
	Encrypt(*kms.EncryptInput) (*kms.EncryptOutput, error)
}
