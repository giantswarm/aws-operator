package region

import (
	"context"

	"github.com/giantswarm/microerror"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"

	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v29/controllercontext"
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v29/key"
)

func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	md, err := key.ToMachineDeployment(obj)
	if err != nil {
		return microerror.Mask(err)
	}
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	// Fetch the cluster for region information.
	var cl v1alpha1.Cluster
	{
		m, err := r.cmaClient.ClusterV1alpha1().Clusters(md.Namespace).Get(key.ClusterID(&md), metav1.GetOptions{})
		if err != nil {
			return microerror.Mask(err)
		}

		cl = *m
	}

	// Simply put the region into the controller context for later use in for
	// instance the tcnp resource.
	{
		cc.Status.TenantCluster.AWS.Region = key.Region(cl)
	}

	return nil
}
