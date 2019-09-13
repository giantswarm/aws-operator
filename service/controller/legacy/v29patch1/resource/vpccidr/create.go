package vpccidr

import (
	"context"

	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/service/controller/legacy/v29patch1/key"
)

func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	cr, err := key.ToCustomObject(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	err = r.addVPCCIDRToContext(ctx, cr)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}
