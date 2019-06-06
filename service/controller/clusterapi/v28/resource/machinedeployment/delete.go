package machinedeployment

import (
	"context"
	"fmt"

	"github.com/giantswarm/microerror"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v28/controllercontext"
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v28/key"
)

func (r *Resource) EnsureDeleted(ctx context.Context, obj interface{}) error {
	cr, err := key.ToCluster(obj)
	if err != nil {
		return microerror.Mask(err)
	}
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "finding machine deployment for cluster")

		in := metav1.ListOptions{
			LabelSelector: fmt.Sprintf("%s=%s", key.LabelCluster, key.ClusterID(cr)),
		}

		out, err := r.cmaClient.ClusterV1alpha1().MachineDeployments(metav1.NamespaceAll).List(in)
		if err != nil {
			return microerror.Mask(err)
		}
		if len(out.Items) == 0 {
			r.logger.LogCtx(ctx, "level", "debug", "message", "did not find machine deployment for cluster")
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling reconciliation")
			return nil
		}
		if len(out.Items) != 1 {
			return microerror.Maskf(executionFailedError, "expected 1 machine deployment got %d", len(out.Items))
		}

		cc.Status.TenantCluster.TCCP.MachineDeployment = out.Items[0]

		r.logger.LogCtx(ctx, "level", "debug", "message", "found machine deployment for cluster")
	}

	return nil
}
