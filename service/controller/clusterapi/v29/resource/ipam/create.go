package ipam

import (
	"context"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/controller/context/reconciliationcanceledcontext"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	cmav1alpha1 "sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"

	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v29/key"
	"github.com/giantswarm/aws-operator/service/network"
)

func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	var cr cmav1alpha1.Cluster
	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "fetching latest Cluster CR")

		oldObj, err := key.ToCluster(obj)
		if err != nil {
			return microerror.Mask(err)
		}

		newObj, err := r.cmaClient.ClusterV1alpha1().Clusters(oldObj.GetNamespace()).Get(oldObj.GetName(), metav1.GetOptions{})
		if err != nil {
			return microerror.Mask(err)
		}
		cr = *newObj

		r.logger.LogCtx(ctx, "level", "debug", "message", "fetched latest Cluster CR")
	}

	if key.StatusClusterNetworkCIDR(cr) == "" {
		r.logger.LogCtx(ctx, "level", "debug", "message", "allocating subnet for Cluster CR")

		callbacks := network.Callbacks{
			Collect: r.collector.Collect,
			Persist: r.persister.NewPersistFunc(ctx, obj),
		}

		_, err := r.allocator.Allocate(ctx, r.networkRange, r.allocatedSubnetMask, callbacks)
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "allocated subnet for Cluster CR")

		r.logger.LogCtx(ctx, "level", "debug", "message", "canceling reconciliation")
		reconciliationcanceledcontext.SetCanceled(ctx)

	} else {
		r.logger.LogCtx(ctx, "level", "debug", "message", "not allocating subnet for Cluster CR")
	}

	return nil
}
