package s3bucket

import (
	"context"

	"github.com/giantswarm/aws-operator/service/controller/v8/key"
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

	bucketState := BucketState{
		Name: key.BucketName(customObject, accountID),
	}

	return bucketState, nil
}
