package cloudformationv1

import (
	"context"

	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/service/keyv1"
)

func (r *Resource) GetDesiredState(ctx context.Context, obj interface{}) (interface{}, error) {
	customObject, err := keyv1.ToCustomObject(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	mainStack, err := newMainStack(customObject)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return mainStack, nil
}
