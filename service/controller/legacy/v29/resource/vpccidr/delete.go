package vpccidr

import (
	"context"

	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/service/controller/legacy/v28/key"
)

func (r *Resource) EnsureDeleted(ctx context.Context, obj interface{}) error {
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
