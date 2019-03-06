package s3bucket

import (
	"context"

	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/service/controller/v23patch1/controllercontext"
	"github.com/giantswarm/aws-operator/service/controller/v23patch1/key"
)

func (r *Resource) GetDesiredState(ctx context.Context, obj interface{}) (interface{}, error) {
	customObject, err := key.ToCustomObject(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	// First bucket must be the delivery log bucket because otherwise
	// other buckets can not forward logs to it
	bucketsState := []BucketState{
		{
			Name:             key.TargetLogBucketName(customObject),
			IsLoggingBucket:  true,
			IsLoggingEnabled: true,
		},
		{
			Name:             key.BucketName(customObject, cc.Status.Cluster.AWSAccount.ID),
			IsLoggingBucket:  false,
			IsLoggingEnabled: true,
		},
	}

	return bucketsState, nil
}
