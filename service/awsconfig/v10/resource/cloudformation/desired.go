package cloudformation

import (
	"context"

	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/service/awsconfig/v10/key"
)

func (r *Resource) GetDesiredState(ctx context.Context, obj interface{}) (interface{}, error) {
	customObject, err := key.ToCustomObject(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var mainStack StackState
	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "computing desired state for the guest cluster main stack")

		mainStack, err = newMainStack(customObject)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "computed desired state for the guest cluster main stack")
	}

	return mainStack, nil
}
