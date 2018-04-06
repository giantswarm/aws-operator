package s3bucket

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/giantswarm/aws-operator/service/awsconfig/v10/key"
	"github.com/giantswarm/microerror"
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

	bucketNames := []string{
		key.TargetLogBucketName(customObject),
		key.BucketName(customObject, accountID),
	}

	bucketsState := []BucketState{}

	for _, bucketName := range bucketNames {
		headInput := &s3.HeadBucketInput{
			Bucket: aws.String(bucketName),
		}
		_, err = r.clients.S3.HeadBucket(headInput)
		if err != nil {
			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("did not find the S3 bucket %q", bucketName))
		}
		if err == nil {
			bucketState := BucketState{
				Name: bucketName,
			}
			bucketsState = append(bucketsState, bucketState)
		}
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", "found the S3 buckets")

	return bucketsState, nil
}
