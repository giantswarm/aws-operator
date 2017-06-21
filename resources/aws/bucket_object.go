package aws

import (
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	microerror "github.com/giantswarm/microkit/error"
)

type BucketObject struct {
	Name   string
	Data   string
	Bucket *Bucket
	AWSEntity
}

// CreateIfNotExists is not implemeted because S3 overwrites bucket objects in
// case of name clashes. This means that a newer CloudConfig with the same name
// as an old one will always overwrite it.
// This shouldn't be a problem, since we use the hash of the CloudConfig
// contents in its name.
// If we decide to use the S3 bucket for other types of files, we might have to
// revisit this.
func (bo *BucketObject) CreateIfNotExists() (bool, error) {
	return false, microerror.MaskAny(notImplementedMethodError)
}

func (bo *BucketObject) CreateOrFail() error {
	if bo.Bucket == nil {
		return microerror.MaskAny(noBucketInBucketObjectError)
	}

	if _, err := bo.Clients.S3.PutObject(&s3.PutObjectInput{
		Body:          strings.NewReader(bo.Data),
		Bucket:        aws.String(bo.Bucket.Name),
		Key:           aws.String(bo.Name),
		ContentLength: aws.Int64(int64(len(bo.Data))),
	}); err != nil {
		return microerror.MaskAny(err)
	}

	return nil
}

func (bo *BucketObject) Delete() error {
	if _, err := bo.Clients.S3.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(bo.Bucket.Name),
		Key:    aws.String(bo.Name),
	}); err != nil {
		return microerror.MaskAny(err)
	}

	return nil
}
