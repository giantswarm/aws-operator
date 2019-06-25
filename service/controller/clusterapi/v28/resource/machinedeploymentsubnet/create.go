package machinedeploymentsubnet

import (
	"context"
	"fmt"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/controller/context/reconciliationcanceledcontext"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/aws-operator/pkg/annotation"
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v28/key"
)

func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	cr, err := key.ToMachineDeployment(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	var cs string
	var ms string
	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "finding cluster for machine deployment")

		cl, err := r.cmaClient.ClusterV1alpha1().Clusters(cr.Namespace).Get(key.WorkerClusterID(cr), metav1.GetOptions{})
		if err != nil {
			return microerror.Mask(err)
		}
		cs = key.StatusClusterNetworkCIDR(*cl)
		ms = cr.Annotations[annotation.MachineDeploymentSubnet]

		r.logger.LogCtx(ctx, "level", "debug", "message", "found cluster for machine deployment")
	}

	{
		if ms != "" {
			r.logger.LogCtx(ctx, "level", "debug", "message", "subnet found in machine deployment")
			r.logger.LogCtx(ctx, "level", "debug", "message", "not updating subnet annotation of machine deployment")
			return nil
		}

		if cs == "" {
			r.logger.LogCtx(ctx, "level", "debug", "message", "subnet not found in cluster")
			r.logger.LogCtx(ctx, "level", "debug", "message", "not updating subnet annotation of machine deployment")
			return nil
		}
	}

	{
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("updating subnet annotation of machine deployment with %q", cs))

		md, err := r.cmaClient.ClusterV1alpha1().MachineDeployments(cr.Namespace).Get(cr.Name, metav1.GetOptions{})
		if err != nil {
			return microerror.Mask(err)
		}

		md.Annotations[annotation.MachineDeploymentSubnet] = cs

		_, err = r.cmaClient.ClusterV1alpha1().MachineDeployments(md.Namespace).Update(md)
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("updated subnet annotation of machine deployment with %q", cs))

		r.logger.LogCtx(ctx, "level", "debug", "message", "canceling reconciliation")
		reconciliationcanceledcontext.SetCanceled(ctx)
	}

	return nil
}
