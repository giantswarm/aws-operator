package cleanupmachinedeployments

import (
	"context"
	"fmt"

	"github.com/giantswarm/microerror"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"

	"github.com/giantswarm/aws-operator/pkg/label"
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v30/key"
)

func (r *Resource) EnsureDeleted(ctx context.Context, obj interface{}) error {
	cr, err := key.ToCluster(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	var machineDeployments []v1alpha1.MachineDeployment
	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "finding MachineDeployments for tenant cluster")

		l := metav1.AddLabelToSelector(
			&metav1.LabelSelector{},
			label.Cluster,
			key.ClusterID(&cr),
		)
		o := metav1.ListOptions{
			LabelSelector: labels.Set(l.MatchLabels).String(),
		}

		list, err := r.cmaClient.ClusterV1alpha1().MachineDeployments(cr.Namespace).List(o)
		if err != nil {
			return microerror.Mask(err)
		}

		machineDeployments = list.Items

		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("found %d MachineDeployments for tenant cluster", len(machineDeployments)))
	}

	if len(machineDeployments) > 0 {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("deleting %d machinedeployments for tenant cluster %#q", len(machineDeployments), key.ClusterID(&cr)))

		var deleted int
		for _, md := range machineDeployments {
			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("deleting machinedeployment %q for tenant cluster %#q", string(md.Namespace+"/"+md.Name), key.ClusterID(&cr)))

			err = r.cmaClient.ClusterV1alpha1().MachineDeployments(md.Namespace).Delete(md.Name, &metav1.DeleteOptions{})
			if err != nil {
				return microerror.Mask(err)
			}

			deleted++

			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("deleted machinedeployment %q for tenant cluster %#q", string(md.Namespace+"/"+md.Name), key.ClusterID(&cr)))
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("deleted %d machinedeployments for tenant cluster %#q", deleted, key.ClusterID(&cr)))
	} else {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("no machinedeployments to be deleted for tenant cluster %#q", key.ClusterID(&cr)))
	}

	return nil
}
