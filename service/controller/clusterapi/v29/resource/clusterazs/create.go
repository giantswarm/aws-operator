package clusterazs

import (
	"context"
	"fmt"

	"github.com/giantswarm/microerror"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	clusterv1alpha1 "sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"

	"github.com/giantswarm/aws-operator/pkg/label"
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v29/controllercontext"
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v29/key"
)

func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	cr, err := key.ToCluster(obj)
	if err != nil {
		return microerror.Mask(err)
	}
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	var machineDeployments []clusterv1alpha1.MachineDeployment
	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "finding MachineDeployments for tenant cluster")

		l := metav1.AddLabelToSelector(
			&v1.LabelSelector{},
			label.Cluster,
			key.ClusterID(cr),
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

	var azs []string
	{
		var azsMap map[string]struct{}

		// Include master's AZ.
		azsMap[key.MasterAvailabilityZone(cr)] = struct{}{}

		// ...and workers'.
		for _, md := range machineDeployments {
			for _, az := range key.WorkerAvailabilityZones(md) {
				azsMap[az] = struct{}{}
			}
		}

		for az, _ := range azsMap {
			azs = append(azs, az)
		}
	}

	{
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("setting cluster availability zones to controllercontext: %#v", azs))

		cc.Status.TenantCluster.AvailabilityZones = azs
	}

	return nil
}
