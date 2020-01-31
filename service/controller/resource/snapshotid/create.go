package snapshotid

import (
	"context"

	"github.com/giantswarm/api/pkg/key"
	"github.com/giantswarm/microerror"
	"k8s.io/apimachinery/pkg/api/meta"

	"github.com/giantswarm/aws-operator/service/controller/controllercontext"
)

func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	cr, err := meta.Accessor(obj)
	if err != nil {
		return microerror.Mask(err)
	}
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	clusterID := key.ClusterID(cr)

	// TODO define key function

	// TODO lookup snapshot

	// TODO define structure
	cc.Status.TenantCluster.TCNP.ASG.MinSize = minSize

	return nil
}
