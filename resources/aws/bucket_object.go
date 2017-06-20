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
