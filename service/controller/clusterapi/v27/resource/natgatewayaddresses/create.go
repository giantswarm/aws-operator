package natgatewayaddresses

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

	err = r.addNATGatewayAddressesToContext(ctx, cr)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}
