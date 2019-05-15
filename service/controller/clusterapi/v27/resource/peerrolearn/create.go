package peerrolearn

import (
	"context"

	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v27/key"
)

func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	cr, err := key.ToCluster(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "finding control plane peer role arn")

		err = r.addPeerRoleARNToContext(ctx, cr)
		if IsNotFound(err) {
			r.logger.LogCtx(ctx, "level", "debug", "message", "did not find control plane peer role arn")
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
			return nil
		} else if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "found control plane peer role arn")
	}

	return nil
}
