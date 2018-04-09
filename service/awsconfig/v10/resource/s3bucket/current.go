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

	bucketState := []BucketState{
		BucketState{
			Name:           key.TargetLogBucketName(customObject),
			IsDeliveryLog:  true,
			LoggingEnabled: true,
		},
		BucketState{
			Name:           key.BucketName(customObject, accountID),
			IsDeliveryLog:  false,
			LoggingEnabled: true,
		},
	}

	currentBucketState := []BucketState{}

	for _, inputBucket := range bucketState {
		headInput := &s3.HeadBucketInput{
			Bucket: aws.String(inputBucket.Name),
		}
		_, err = r.clients.S3.HeadBucket(headInput)
		if IsBucketNotFound(err) {
			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("did not find the S3 bucket %q", inputBucket.Name))
		}
		if err != nil {
			return []BucketState{}, nil
		}
		if err == nil {
			currentBucketState = append(currentBucketState, inputBucket)
		}
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", "found the S3 buckets")

	return currentBucketState, nil
}
