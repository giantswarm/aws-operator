package aws

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/giantswarm/microerror"
)

type Bucket struct {
	Name string
	AWSEntity
}

func (b *Bucket) CreateIfNotExists() (bool, error) {
	if err := b.CreateOrFail(); err != nil {
		underlying := microerror.Cause(err)
		if awserr, ok := underlying.(awserr.Error); ok {
			if awserr.Code() == s3.ErrCodeBucketAlreadyOwnedByYou {
				return false, nil
			}
		}

		return false, microerror.Mask(err)
	}
	return true, nil
}

func (b *Bucket) CreateOrFail() error {
	if _, err := b.Clients.S3.CreateBucket(&s3.CreateBucketInput{
		Bucket: aws.String(b.Name),
	}); err != nil {
		return microerror.Mask(err)
	}

	if err := b.Clients.S3.WaitUntilBucketExists(&s3.HeadBucketInput{
		Bucket: aws.String(b.Name),
	}); err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func (b *Bucket) Delete() error {
	// List bucket objects and delete them one by one.
	objects, err := b.Clients.S3.ListObjects(&s3.ListObjectsInput{
		Bucket: aws.String(b.Name),
	})
	if err != nil {
		return microerror.Mask(err)
	}

	for _, o := range objects.Contents {
		if _, err := b.Clients.S3.DeleteObject(&s3.DeleteObjectInput{
			Bucket: aws.String(b.Name),
			Key:    aws.String(*o.Key),
		}); err != nil {
			return microerror.Mask(err)
		}
	}

	if _, err := b.Clients.S3.DeleteBucket(&s3.DeleteBucketInput{
		Bucket: aws.String(b.Name),
	}); err != nil {
		return microerror.Mask(err)
	}

	return nil
}
