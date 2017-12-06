package s3bucketv2

import (
	"context"

	"github.com/giantswarm/aws-operator/service/keyv2"
	"github.com/giantswarm/microerror"
)

func (r *Resource) GetDesiredState(ctx context.Context, obj interface{}) (interface{}, error) {
	customObject, err := keyv2.ToCustomObject(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	accountID, err := r.awsService.GetAccountID()
	if err != nil {
		return nil, microerror.Mask(err)
	}

	bucketState := BucketState{
		Name: keyv2.BucketName(customObject, accountID),
	}

	return bucketState, nil
}
