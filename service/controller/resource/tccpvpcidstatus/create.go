package tccpvpcidstatus

import (
	"context"
	"fmt"

	infrastructurev1alpha2 "github.com/giantswarm/apiextensions/v2/pkg/apis/infrastructure/v1alpha2"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/v2/pkg/controller/context/reconciliationcanceledcontext"
	"k8s.io/apimachinery/pkg/types"

	"github.com/giantswarm/aws-operator/service/controller/controllercontext"
	"github.com/giantswarm/aws-operator/service/controller/key"
)

func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	var cr infrastructurev1alpha2.AWSCluster
	{
		cl, err := key.ToCluster(ctx, obj)
		if err != nil {
			return microerror.Mask(err)
		}

		err = r.k8sClient.CtrlClient().Get(ctx, types.NamespacedName{Name: cl.Name, Namespace: cl.Namespace}, &cr)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	if cr.Status.Provider.Network.VPCID != "" {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("cluster %#q already has vpc id in status", cr.GetName()))
		r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")

		return nil
	}

	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "updating cluster status with vpc id")

		cr.Status.Provider.Network.VPCID = cc.Status.TenantCluster.TCCP.VPC.ID

		err := r.k8sClient.CtrlClient().Status().Update(ctx, &cr)
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "updated cluster status with vpc id")
		r.logger.LogCtx(ctx, "level", "debug", "message", "canceling reconciliation")
		reconciliationcanceledcontext.SetCanceled(ctx)
	}

	return nil
}
