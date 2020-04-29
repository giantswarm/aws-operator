package s3object

import (
	"bytes"
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/controller/context/resourcecanceledcontext"

	"github.com/giantswarm/aws-operator/service/controller/controllercontext"
	"github.com/giantswarm/aws-operator/service/controller/key"
)

func (r *Resource) GetCurrentState(ctx context.Context, obj interface{}) (interface{}, error) {
	customObject, err := key.ToCustomObject(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	// During deletion, it might happen that the encryption key got already
	// deleted. In such a case we do not have to do anything here anymore. The
	// desired state computation usually requires the encryption key to come up
	// with the deletion state, but in case it is gone we do not have to do
	// anything here anymore. The current implementation relies on the bucket
	// deletion of the s3bucket resource, which deletes all S3 objects and the
	// bucket itself.
	if key.IsDeleted(customObject) {
		if cc.Status.TenantCluster.Encryption.Key == "" {
			r.logger.LogCtx(ctx, "level", "debug", "message", "no encryption key in controller context")

			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
			resourcecanceledcontext.SetCanceled(ctx)

			return nil, nil
		}
	}

	bucketName := key.BucketName(customObject, cc.Status.TenantCluster.AWSAccountID)

	var objects []*s3.Object
	{
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("finding the S3 bucket %#q", bucketName))

		i := &s3.ListObjectsV2Input{
			Bucket: aws.String(bucketName),
		}

		o, err := cc.Client.TenantCluster.AWS.S3.ListObjectsV2(i)
		if IsBucketNotFound(err) {
			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("did not find the S3 bucket %#q", bucketName))
			return nil, nil
		} else if err != nil {
			return nil, microerror.Mask(err)
		}

		objects = o.Contents

		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("found the S3 bucket %#q", bucketName))
	}

	currentBucketState := map[string]BucketObjectState{}
	{
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("finding the contents of %d S3 objects", len(objects)))

		for _, object := range objects {
			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("finding the content of the S3 object %#q", *object.Key))

			currentBucketState[*object.Key], err = r.getBucketObject(ctx, bucketName, *object.Key)
			if err != nil {
				return nil, microerror.Mask(err)
			}

			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("found the content of the S3 object %#q", *object.Key))
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("found the contents of %d S3 objects", len(objects)))
	}

	return currentBucketState, nil
}

func (r *Resource) getBucketObject(ctx context.Context, bucketName string, keyName string) (BucketObjectState, error) {
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return BucketObjectState{}, microerror.Mask(err)
	}
	input := &s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(keyName),
	}
	result, err := cc.Client.TenantCluster.AWS.S3.GetObject(input)
	if IsObjectNotFound(err) || IsBucketNotFound(err) {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("did not find S3 object %#q", keyName))
		return BucketObjectState{}, nil
	} else if err != nil {
		return BucketObjectState{}, microerror.Mask(err)
	}

	var body string
	{
		buf := new(bytes.Buffer)
		_, err = buf.ReadFrom(result.Body)
		if err != nil {
			return BucketObjectState{}, microerror.Mask(err)
		}
		body = buf.String()
	}

	decrypted, err := r.cloudConfig.DecryptTemplate(ctx, body)
	if err != nil {
		return BucketObjectState{}, microerror.Mask(err)
	}

	object := BucketObjectState{
		Bucket: bucketName,
		Body:   body,
		Hash:   hashIgnition(decrypted),
		Key:    keyName,
	}

	return object, nil
}
