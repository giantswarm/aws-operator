package cloudformation

import (
	"context"

	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/service/key"
)

func (r *Resource) GetDesiredState(ctx context.Context, obj interface{}) (interface{}, error) {
	customObject, err := key.ToCustomObject(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	// no-op if we are not using cloudformation
	if !key.UseCloudFormation(customObject) {
		r.logger.LogCtx(ctx, "debug", "not processing cloudformation stack")
		return nil, nil
	}

	mainStack, err := newMainStack(customObject)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return mainStack, nil
}
