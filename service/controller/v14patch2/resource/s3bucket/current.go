package s3bucket

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/service/controller/v14patch2/controllercontext"
	"github.com/giantswarm/aws-operator/service/controller/v14patch2/key"
)

func (r *Resource) GetCurrentState(ctx context.Context, obj interface{}) (interface{}, error) {
	customObject, err := key.ToCustomObject(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", "looking for the S3 buckets")

	sc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	accountID, err := sc.AWSService.GetAccountID()
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

		isCreated, err := r.isBucketCreated(ctx, inputBucketName)
		if err != nil {
			return nil, microerror.Mask(err)
		}
		if !isCreated {
			continue
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("find the S3 bucket %q", inputBucketName))

		lc, err := r.getLoggingConfiguration(ctx, inputBucketName)
		if err != nil {
			return nil, microerror.Mask(err)
		}
		inputBucket.LoggingEnabled = isLoggingEnabled(lc)
		inputBucket.IsLoggingBucket = isLoggingBucket(inputBucketName, lc)

		currentBucketState = append(currentBucketState, inputBucket)
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", "found the S3 buckets")

	return currentBucketState, nil
}

func (r *Resource) isBucketCreated(ctx context.Context, name string) (bool, error) {
	sc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return false, microerror.Mask(err)
	}

	headInput := &s3.HeadBucketInput{
		Bucket: aws.String(name),
	}
	_, err = sc.AWSClient.S3.HeadBucket(headInput)
	if IsBucketNotFound(err) {
		return false, nil
	} else if err != nil {
		return false, microerror.Mask(err)
	}

	return true, nil
}

func (r *Resource) getLoggingConfiguration(ctx context.Context, name string) (*s3.GetBucketLoggingOutput, error) {
	sc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	bucketLoggingInput := &s3.GetBucketLoggingInput{
		Bucket: aws.String(name),
	}
	bucketLoggingOutput, err := sc.AWSClient.S3.GetBucketLogging(bucketLoggingInput)
	if err != nil {
		return bucketLoggingOutput, microerror.Mask(err)
	}

	return bucketLoggingOutput, nil
}

func isLoggingEnabled(lc *s3.GetBucketLoggingOutput) bool {
	if lc.LoggingEnabled != nil {
		return true
	}

	return false
}

func isLoggingBucket(name string, lc *s3.GetBucketLoggingOutput) bool {
	if lc.LoggingEnabled != nil {
		if *lc.LoggingEnabled.TargetBucket == name {
			return true
		}
	}

	return false
}
