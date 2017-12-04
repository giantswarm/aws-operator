package s3bucketv1

import (
	"context"

	"github.com/giantswarm/aws-operator/service/key"
	"github.com/giantswarm/microerror"
)

func (r *Resource) GetDesiredState(ctx context.Context, obj interface{}) (interface{}, error) {
	customObject, err := key.ToCustomObject(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	bucketState := BucketState{
		Name: key.BucketName(customObject, r.awsConfig.AccountID()),
	}

	return bucketState, nil
}
