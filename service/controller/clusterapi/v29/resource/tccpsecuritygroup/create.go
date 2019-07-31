package tccpsecuritygroup

import (
	"context"

	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v29/key"
)

func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	cr, err := key.ToMachineDeployment(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	err = r.addInfoToCtx(ctx, cr)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}
