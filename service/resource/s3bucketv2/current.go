package s3bucketv2

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/giantswarm/aws-operator/service/keyv2"
	"github.com/giantswarm/microerror"
)

func (r *Resource) GetCurrentState(ctx context.Context, obj interface{}) (interface{}, error) {
	customObject, err := keyv2.ToCustomObject(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "debug", "looking for the S3 bucket")

	accountID, err := r.awsService.GetAccountID()
	if err != nil {
		return nil, microerror.Mask(err)
	}

	bucketName := keyv2.BucketName(customObject, accountID)
	headInput := &s3.HeadBucketInput{
		Bucket: aws.String(bucketName),
	}
	_, err = r.clients.S3.HeadBucket(headInput)
	if IsBucketNotFound(err) {
		r.logger.LogCtx(ctx, "debug", "did not find the S3 bucket")
		return BucketState{}, nil
	}
	if err != nil {
		return BucketState{}, microerror.Mask(err)
	}

	bucketState := BucketState{
		Name: bucketName,
	}

	r.logger.LogCtx(ctx, "debug", "found the S3 bucket")

	return bucketState, nil
}
