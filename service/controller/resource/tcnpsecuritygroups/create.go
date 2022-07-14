package tcnpsecuritygroups

import (
	"context"

	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/v12/service/controller/key"
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
