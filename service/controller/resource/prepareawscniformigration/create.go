package prepareawscniformigration

import (
	"context"

	"github.com/giantswarm/k8smetadata/pkg/annotation"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/v7/pkg/controller/context/reconciliationcanceledcontext"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	apiv1beta1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/aws-operator/v13/service/controller/controllercontext"
	"github.com/giantswarm/aws-operator/v13/service/controller/key"
)

const (
	dsNamespace = "kube-system"
	dsName      = "aws-node"
	envVarName  = "AWS_VPC_K8S_CNI_EXCLUDE_SNAT_CIDRS"
)

func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	var err error
	cr, err := key.ToCluster(ctx, obj)
	if err != nil {
		return microerror.Mask(err)
	}

	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	if cc.Client.TenantCluster.K8s == nil {
		r.logger.Debugf(ctx, "kubernetes clients are not available in controller context yet")
		r.logger.Debugf(ctx, "canceling resource")

		return nil
	}

	// Only run this if the Cluster CR has the cilium pod annotation
	cluster := apiv1beta1.Cluster{}
	err = r.ctrlClient.Get(ctx, client.ObjectKey{Namespace: cr.Namespace, Name: key.ClusterID(&cr)}, &cluster)
	if err != nil {
		return microerror.Mask(err)
	}

	if key.CiliumPodsCIDRBlock(cluster) == "" {
		r.logger.Debugf(ctx, "Cluster CR has no %q annotation, nothing to do", annotation.CiliumPodCidr)
	}

	wcCtrlClient := cc.Client.TenantCluster.K8s.CtrlClient()

	ds := &v1.DaemonSet{}
	err = wcCtrlClient.Get(ctx, client.ObjectKey{Name: dsName, Namespace: dsNamespace}, ds)
	if apierrors.IsNotFound(err) {
		// All good.
		r.logger.Debugf(ctx, "Daemonset %q was not found in namespace %q", dsName, dsNamespace)
		return nil
	} else if err != nil {
		return microerror.Mask(err)
	}

	// Ensure aws-node daemonset has needed env var.
	found := false
	for _, env := range ds.Spec.Template.Spec.Containers[0].Env {
		if env.Name == envVarName {
			if env.Value == key.CiliumPodsCIDRBlock(cluster) {
				// Env var found and correct. Check if daemonset is updated.

				if ds.Status.DesiredNumberScheduled != ds.Status.CurrentNumberScheduled {
					r.logger.Debugf(ctx, "Daemonset %q has needed env var %q but not all replicas are healthy. Canceling reconciliation", dsName)
					reconciliationcanceledcontext.SetCanceled(ctx)
				}

				return nil
			} else {
				env.Value = key.CiliumPodsCIDRBlock(cluster)
				r.logger.Debugf(ctx, "Daemonset %q has needed env var %q but value is wrong", dsName, envVarName)
				found = true
				break
			}
		}
	}

	if !found {
		// Add env var.
		r.logger.Debugf(ctx, "Daemonset %q doesn't have needed env var %q", dsName, envVarName)
		ds.Spec.Template.Spec.Containers[0].Env = append(ds.Spec.Template.Spec.Containers[0].Env, corev1.EnvVar{
			Name:  envVarName,
			Value: key.CiliumPodsCIDRBlock(cluster),
		})
	}

	err = wcCtrlClient.Update(ctx, ds)
	if err != nil {
		return microerror.Mask(err)
	}

	// Wait for next reconciliation loop.
	r.logger.Debugf(ctx, "Daemonset %q was updated, canceling reconciliation", dsName)
	reconciliationcanceledcontext.SetCanceled(ctx)

	return nil
}
