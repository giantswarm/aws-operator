package s3bucketv1

import (
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/s3"
)

// BucketState is the state representation on which the resource methods work.
type BucketState struct {
	Name string
}

type Clients struct {
	IAM IAMClient
	S3  S3Client
}

// IAMClient describes the methods required to be implemented by a IAM AWS client.
type IAMClient interface {
	GetUser(*iam.GetUserInput) (*iam.GetUserOutput, error)
}

// S3Client describes the methods required to be implemented by a S3 AWS client.
type S3Client interface {
	HeadBucket(*s3.HeadBucketInput) (*s3.HeadBucketOutput, error)
}
