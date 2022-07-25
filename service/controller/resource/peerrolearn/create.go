package peerrolearn

import (
	"context"

	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/v13/service/controller/key"
)

func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	cr, err := key.ToCluster(ctx, obj)
	if err != nil {
		return microerror.Mask(err)
	}

	{
		r.logger.Debugf(ctx, "finding control plane peer role arn")

		err = r.addPeerRoleARNToContext(ctx, cr)
		if IsNotFound(err) {
			r.logger.Debugf(ctx, "did not find control plane peer role arn")
			r.logger.Debugf(ctx, "canceling resource")
			return nil
		} else if err != nil {
			return microerror.Mask(err)
		}

		r.logger.Debugf(ctx, "found control plane peer role arn")
	}

	return nil
}
