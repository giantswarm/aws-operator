package s3object

import (
	"bytes"
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/controller/context/resourcecanceledcontext"

	"github.com/giantswarm/aws-operator/service/controller/v22/controllercontext"
	"github.com/giantswarm/aws-operator/service/controller/v22/key"
)

func (r *Resource) GetCurrentState(ctx context.Context, obj interface{}) (interface{}, error) {
	customObject, err := key.ToCustomObject(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", "looking for S3 objects")

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
		if cc.Status.Cluster.EncryptionKey == "" {
			r.logger.LogCtx(ctx, "level", "debug", "message", "no encryption key in controller context")

			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
			resourcecanceledcontext.SetCanceled(ctx)

			return nil, nil
		}
	}

	bucketName := key.BucketName(customObject, cc.Status.Cluster.AWSAccount.ID)
	input := &s3.ListObjectsV2Input{
		Bucket: aws.String(bucketName),
	}
	result, err := cc.AWSClient.S3.ListObjectsV2(input)
	// the bucket can be already deleted with all the objects in it, it is ok if so.
	if IsBucketNotFound(err) {
		r.logger.LogCtx(ctx, "level", "debug", "message", "S3 object's bucket not found, no current objects present")
		return nil, nil
	} else if err != nil {
		return nil, microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", "found the S3 objects")

	output := map[string]BucketObjectState{}
	for _, object := range result.Contents {
		body, err := r.getBucketObjectBody(ctx, bucketName, *object.Key)
		if err != nil {
			return output, microerror.Mask(err)
		}

		output[*object.Key] = BucketObjectState{
			Body:   body,
			Bucket: bucketName,
			Key:    *object.Key,
		}
	}

	return output, nil
}

func (r *Resource) getBucketObjectBody(ctx context.Context, bucketName string, keyName string) (string, error) {
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return "", microerror.Mask(err)
	}
	input := &s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(keyName),
	}
	result, err := cc.AWSClient.S3.GetObject(input)
	if IsObjectNotFound(err) || IsBucketNotFound(err) {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("did not find S3 object '%s'", keyName))
		return "", nil
	} else if err != nil {
		return "", microerror.Mask(err)
	}

	var body string
	{
		buf := new(bytes.Buffer)
		buf.ReadFrom(result.Body)
		body = buf.String()
	}

	return body, nil
}
