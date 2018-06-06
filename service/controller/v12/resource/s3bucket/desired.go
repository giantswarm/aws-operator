package s3bucket

import (
	"context"

	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/service/controller/v12/controllercontext"
	"github.com/giantswarm/aws-operator/service/controller/v12/key"
)

func (r *Resource) GetDesiredState(ctx context.Context, obj interface{}) (interface{}, error) {
	customObject, err := key.ToCustomObject(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	sc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	accountID, err := sc.AWSService.GetAccountID()
	if err != nil {
		return nil, microerror.Mask(err)
	}

	// First bucket must be the delivery log bucket because otherwise
	// other buckets can not forward logs to it
	bucketsState := []BucketState{
		{
			Name:            key.TargetLogBucketName(customObject),
			IsLoggingBucket: true,
			LoggingEnabled:  true,
		},
		{
			Name:            key.BucketName(customObject, accountID),
			IsLoggingBucket: false,
			LoggingEnabled:  true,
		},
	}

	return bucketsState, nil
}
