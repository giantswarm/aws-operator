package s3object

import (
	"bytes"
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/giantswarm/aws-operator/service/awsconfig/v10/key"
	"github.com/giantswarm/microerror"
)

func (r *Resource) GetCurrentState(ctx context.Context, obj interface{}) (interface{}, error) {
	output := map[string]BucketObjectState{}
	customObject, err := key.ToCustomObject(obj)
	if err != nil {
		return output, microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", "looking for S3 objects")

	accountID, err := r.awsService.GetAccountID()
	if err != nil {
		return output, microerror.Mask(err)
	}

	bucketName := key.BucketName(customObject, accountID)
	input := &s3.ListObjectsV2Input{
		Bucket: aws.String(bucketName),
	}
	result, err := r.awsClients.S3.ListObjectsV2(input)
	// the bucket can be already deleted with all the objects in it, it is ok if so.
	if IsBucketNotFound(err) {
		r.logger.LogCtx(ctx, "level", "debug", "message", "S3 object's bucket not found, no current objects present")
		return output, nil
	}
	// we don't expect other errors.
	if err != nil {
		return output, microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", "found the S3 objects")

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
	var body string

	input := &s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(keyName),
	}
	result, err := r.awsClients.S3.GetObject(input)
	if IsObjectNotFound(err) || IsBucketNotFound(err) {
		r.logger.LogCtx(ctx, "level", "info", "message", fmt.Sprintf("did not find S3 object '%s'", keyName))

		// fall through
		return body, nil
	}
	if err != nil {
		return body, microerror.Mask(err)
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(result.Body)
	body = buf.String()

	return body, nil
}
