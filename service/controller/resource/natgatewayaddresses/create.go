package natgatewayaddresses

import (
	"context"

	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/service/controller/key"
)

func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
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
