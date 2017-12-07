package cloudformationv2

import (
	"context"

	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/service/keyv2"
)

func (r *Resource) GetDesiredState(ctx context.Context, obj interface{}) (interface{}, error) {
	customObject, err := keyv2.ToCustomObject(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	mainStack, err := newMainStack(customObject)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return mainStack, nil
}
