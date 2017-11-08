package cloudformation

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

	mainStack, err := newMainStack(customObject)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return mainStack, nil
}
