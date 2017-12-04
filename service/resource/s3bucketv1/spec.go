package s3bucketv1

import (
	"github.com/aws/aws-sdk-go/service/s3"
)

// BucketState is the state representation on which the resource methods work.
type BucketState struct {
	Name string
}

type Clients struct {
	S3 S3Client
}

// S3Client describes the methods required to be implemented by a S3 AWS client.
type S3Client interface {
	CreateBucket(*s3.CreateBucketInput) (*s3.CreateBucketOutput, error)
	HeadBucket(*s3.HeadBucketInput) (*s3.HeadBucketOutput, error)
}
