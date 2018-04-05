package s3bucket

import (
	"context"

	"github.com/giantswarm/aws-operator/service/awsconfig/v10/key"
	"github.com/giantswarm/microerror"
)

func (r *Resource) GetDesiredState(ctx context.Context, obj interface{}) (interface{}, error) {
	customObject, err := key.ToCustomObject(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	accountID, err := r.awsService.GetAccountID()
	if err != nil {
		return nil, microerror.Mask(err)
	}

	// First bucket must to be the delivery log bucket because otherwise
	// other buckets can not forward logs to it
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

	return bucketState, nil
}
