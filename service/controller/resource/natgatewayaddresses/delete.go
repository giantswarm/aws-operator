package natgatewayaddresses

import (
	"context"

	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/v13/service/controller/key"
)

func (r *Resource) EnsureDeleted(ctx context.Context, obj interface{}) error {
	cr, err := key.ToCluster(ctx, obj)
	if err != nil {
		return microerror.Mask(err)
	}

	err = r.addNATGatewayAddressesToContext(ctx, cr)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}
