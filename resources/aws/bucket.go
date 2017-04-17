package aws

import (
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"
	microerror "github.com/giantswarm/microkit/error"
	"github.com/juju/errgo"
)

type Bucket struct {
	Name string
	AWSEntity
}

func (b *Bucket) CreateIfNotExists() (bool, error) {
	if err := b.CreateOrFail(); err != nil {
		underlying := errgo.Cause(err)
		if awserr, ok := underlying.(awserr.Error); ok {
			if awserr.Code() == s3.ErrCodeBucketAlreadyOwnedByYou {
				return false, nil
			}
		}

		return false, microerror.MaskAny(err)
	}
	return true, nil
}

func (b *Bucket) CreateOrFail() error {
	if _, err := b.Clients.S3.CreateBucket(&s3.CreateBucketInput{
		Bucket: aws.String(b.Name),
	}); err != nil {
		return microerror.MaskAny(err)
	}

	if err := b.Clients.S3.WaitUntilBucketExists(&s3.HeadBucketInput{
		Bucket: aws.String(b.Name),
	}); err != nil {
		return microerror.MaskAny(err)
	}

	return nil
}

func (b *Bucket) Delete() error {
	if _, err := b.Clients.S3.DeleteBucket(&s3.DeleteBucketInput{
		Bucket: aws.String(b.Name),
	}); err != nil {
		return microerror.MaskAny(err)
	}

	return nil
}

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
