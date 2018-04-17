package s3bucket

import (
	"github.com/aws/aws-sdk-go/service/s3"
)

// BucketState is the state representation on which the resource methods work.
type BucketState struct {
	Name            string
	IsLoggingBucket bool
	LoggingEnabled  bool
}

type Clients struct {
	S3 S3Client
}

// S3Client describes the methods required to be implemented by a S3 AWS client.
type S3Client interface {
	CreateBucket(*s3.CreateBucketInput) (*s3.CreateBucketOutput, error)
	DeleteBucket(*s3.DeleteBucketInput) (*s3.DeleteBucketOutput, error)
	DeleteObject(*s3.DeleteObjectInput) (*s3.DeleteObjectOutput, error)
	DeleteObjects(*s3.DeleteObjectsInput) (*s3.DeleteObjectsOutput, error)
	GetBucketLogging(*s3.GetBucketLoggingInput) (*s3.GetBucketLoggingOutput, error)
	HeadBucket(*s3.HeadBucketInput) (*s3.HeadBucketOutput, error)
	ListObjectsV2(*s3.ListObjectsV2Input) (*s3.ListObjectsV2Output, error)
	PutBucketAcl(*s3.PutBucketAclInput) (*s3.PutBucketAclOutput, error)
	PutBucketLogging(*s3.PutBucketLoggingInput) (*s3.PutBucketLoggingOutput, error)
	PutBucketTagging(*s3.PutBucketTaggingInput) (*s3.PutBucketTaggingOutput, error)
}
