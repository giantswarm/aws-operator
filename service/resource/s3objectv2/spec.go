package s3objectv2

import (
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/certificatetpr"
	"github.com/giantswarm/randomkeytpr"
)

const (
	prefixMaster = "master"
	prefixWorker = "worker"

	encryptionConfigTemplate = `
kind: EncryptionConfig
apiVersion: v1
resources:
  - resources:
    - secrets
    providers:
    - aescbc:
        keys:
        - name: key1
          secret: {{.EncryptionKey}}
    - identity: {}
`
)

type BucketObjectState struct {
	MasterCloudConfig BucketObjectInstance
	WorkerCloudConfig BucketObjectInstance
}

type BucketObjectInstance struct {
	Bucket string
	Body   string
	Key    string
}

type Clients struct {
	S3  S3Client
	KMS KMSClient
}

type S3Client interface {
	GetObject(*s3.GetObjectInput) (*s3.GetObjectOutput, error)
	PutObject(*s3.PutObjectInput) (*s3.PutObjectOutput, error)
	DeleteObject(*s3.DeleteObjectInput) (*s3.DeleteObjectOutput, error)
}

type KMSClient interface {
	Encrypt(*kms.EncryptInput) (*kms.EncryptOutput, error)
}

type AwsService interface {
	GetAccountID() (string, error)
	GetKeyArn(string) (string, error)
}

type CloudConfigService interface {
	NewMasterTemplate(v1alpha1.AWSConfig, certificatetpr.CompactTLSAssets, randomkeytpr.CompactRandomKeyAssets) (string, error)
	NewWorkerTemplate(v1alpha1.AWSConfig, certificatetpr.CompactTLSAssets) (string, error)
}

type CertWatcher interface {
	SearchCerts(string) (certificatetpr.AssetsBundle, error)
}

type RandomKeyWatcher interface {
	SearchKeys(clusterID string) (map[randomkeytpr.Key][]byte, error)
}
