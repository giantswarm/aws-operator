package s3bucket

import (
	"context"

	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v27/controllercontext"
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v27/legacykey"
)

func (r *Resource) GetDesiredState(ctx context.Context, obj interface{}) (interface{}, error) {
	customObject, err := legacykey.ToCustomObject(obj)
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
			Name:             legacykey.TargetLogBucketName(customObject),
			IsLoggingBucket:  true,
			IsLoggingEnabled: true,
		},
		{
			Name:             legacykey.BucketName(customObject, cc.Status.TenantCluster.AWSAccountID),
			IsLoggingBucket:  false,
			IsLoggingEnabled: true,
		},
	}

	return bucketsState, nil
}
