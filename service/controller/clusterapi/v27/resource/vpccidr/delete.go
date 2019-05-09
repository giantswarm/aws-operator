package vpccidr

import (
	"context"

	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v27/key"
)

func (r *Resource) EnsureDeleted(ctx context.Context, obj interface{}) error {
	cluster, err := key.ToCluster(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	err = r.addVPCCIDRToContext(ctx, cluster)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}
