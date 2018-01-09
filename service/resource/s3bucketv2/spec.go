package s3bucketv2

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
	DeleteBucket(*s3.DeleteBucketInput) (*s3.DeleteBucketOutput, error)
	HeadBucket(*s3.HeadBucketInput) (*s3.HeadBucketOutput, error)
	DeleteObjects(*s3.DeleteObjectsInput) (*s3.DeleteObjectsOutput, error)
	ListObjectsV2(*s3.ListObjectsV2Input) (*s3.ListObjectsV2Output, error)
	DeleteObject(*s3.DeleteObjectInput) (*s3.DeleteObjectOutput, error)
}
