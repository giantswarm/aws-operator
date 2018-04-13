package s3bucket

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/service/awsconfig/v10/key"
)

func (r *Resource) GetCurrentState(ctx context.Context, obj interface{}) (interface{}, error) {
	customObject, err := key.ToCustomObject(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", "looking for the S3 buckets")

	accountID, err := r.awsService.GetAccountID()
	if err != nil {
		return nil, microerror.Mask(err)
	}

	bucketStateNames := []string{
		key.TargetLogBucketName(customObject),
		key.BucketName(customObject, accountID),
	}

	var currentBucketState []BucketState
	for _, inputBucketName := range bucketStateNames {
		inputBucket := BucketState{
			Name: inputBucketName,
		}

		headInput := &s3.HeadBucketInput{
			Bucket: aws.String(inputBucketName),
		}
		_, err = r.clients.S3.HeadBucket(headInput)
		if IsBucketNotFound(err) {
			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("did not find the S3 bucket %q", inputBucketName))
		} else if err != nil {
			return nil, microerror.Mask(err)
		}

		bucketLoggingInput := &s3.GetBucketLoggingInput{
			Bucket: aws.String(inputBucketName),
		}
		bucketLoggingOutput, err := r.clients.S3.GetBucketLogging(bucketLoggingInput)
		if IsBucketNotFound(err) {
			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("did not find the S3 bucket logging for %q", inputBucketName))
		} else if err != nil {
			return nil, microerror.Mask(err)
		}
		if bucketLoggingOutput.LoggingEnabled != nil {
			inputBucket.LoggingEnabled = true
			if *bucketLoggingOutput.LoggingEnabled.TargetBucket == inputBucketName {
				inputBucket.IsLoggingBucket = true
			}
		}
		currentBucketState = append(currentBucketState, inputBucket)
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", "found the S3 buckets")

	return currentBucketState, nil
}
