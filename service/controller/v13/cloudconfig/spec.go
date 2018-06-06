package cloudconfig

import (
	"github.com/aws/aws-sdk-go/service/kms"
)

const (
	// CloudConfigVersion defines the version of k8scloudconfig in use.
	// It is used in the main stack output and S3 object paths.
	CloudConfigVersion = "v_3_3_2"
)

type KMSClient interface {
	Encrypt(*kms.EncryptInput) (*kms.EncryptOutput, error)
}
