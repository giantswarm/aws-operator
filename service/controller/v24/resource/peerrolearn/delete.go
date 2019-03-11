package peerrolearn

import (
	"context"

	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/service/controller/v24/key"
)

func (r *Resource) EnsureDeleted(ctx context.Context, obj interface{}) error {
	cr, err := key.ToCustomObject(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	err = r.addPeerRoleARNToContext(ctx, cr)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}
