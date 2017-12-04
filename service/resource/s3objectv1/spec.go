package s3objectv1

import "github.com/aws/aws-sdk-go/service/s3"

const (
	prefixWorker = "worker"
)

type BucketObjectState struct {
	WorkerCloudConfig BucketObjectInstance
}

type BucketObjectInstance struct {
	Bucket string
	Body   string
	Key    string
}

type Clients struct {
	S3 S3Client
}

type S3Client interface {
	GetObject(*s3.GetObjectInput) (*s3.GetObjectOutput, error)
	PutObject(*s3.PutObjectInput) (*s3.PutObjectOutput, error)
	DeleteObject(*s3.DeleteObjectInput) (*s3.DeleteObjectOutput, error)
}

type AwsService interface {
	GetAccountID() (string, error)
}
