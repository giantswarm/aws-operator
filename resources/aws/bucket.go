package aws

import (
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	microerror "github.com/giantswarm/microkit/error"

	awsclient "github.com/giantswarm/aws-operator/client/aws"
)

type Bucket struct {
	Name string
	AWSEntity
}

func (b *Bucket) CreateIfNotExists() (bool, error) {
	if err := b.CreateOrFail(); err != nil {
		if strings.Contains(err.Error(), awsclient.BucketAlreadyExists) || strings.Contains(err.Error(), awsclient.BucketAlreadyOwnedByYou) {
			return false, nil
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
