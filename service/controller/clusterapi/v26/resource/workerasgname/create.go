package workerasgname

import (
	"context"

	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/controller/context/reconciliationcanceledcontext"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v26/controllercontext"
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v26/legacykey"
)

func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	var customObject v1alpha1.AWSConfig
	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "finding latest version of custom resource")

		oldObj, err := legacykey.ToCustomObject(obj)
		if err != nil {
			return microerror.Mask(err)
		}

		newObj, err := r.g8sClient.ProviderV1alpha1().AWSConfigs(oldObj.GetNamespace()).Get(oldObj.GetName(), metav1.GetOptions{})
		if err != nil {
			return microerror.Mask(err)
		}
		customObject = *newObj

		r.logger.LogCtx(ctx, "level", "debug", "message", "found latest version of custom resource")
	}

	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "finding the tenant cluster worker ASG name in the CR")

		if customObject.Status.AWS.AutoScalingGroup.Name != "" {
			cc.Status.TenantCluster.TCCP.ASG.Name = customObject.Status.AWS.AutoScalingGroup.Name

			r.logger.LogCtx(ctx, "level", "debug", "message", "found the tenant cluster worker ASG name in the CR")
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
			return nil
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "did not find the tenant cluster worker ASG name in the CR")
	}

	if cc.Status.TenantCluster.TCCP.ASG.Name != "" {
		r.logger.LogCtx(ctx, "level", "debug", "message", "updating CR status")

		customObject.Status.AWS.AutoScalingGroup.Name = cc.Status.TenantCluster.TCCP.ASG.Name

		_, err = r.g8sClient.ProviderV1alpha1().AWSConfigs(customObject.Namespace).UpdateStatus(&customObject)
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "updated CR status")
		r.logger.LogCtx(ctx, "level", "debug", "message", "canceling reconciliation")
		reconciliationcanceledcontext.SetCanceled(ctx)
	}

	return nil
}
