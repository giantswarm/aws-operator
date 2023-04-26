package awscnicleaner

import (
	"context"
	"fmt"
	"time"

	"github.com/giantswarm/backoff"
	"github.com/giantswarm/k8smetadata/pkg/annotation"
	"github.com/giantswarm/microerror"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	apiv1beta1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	annotation2 "github.com/giantswarm/aws-operator/v14/pkg/annotation"
	"github.com/giantswarm/aws-operator/v14/service/controller/controllercontext"
	"github.com/giantswarm/aws-operator/v14/service/controller/key"
)

const (
	dsNamespace = "kube-system"
	dsName      = "aws-node"
)

func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	var err error
	cr, err := key.ToCluster(ctx, obj)
	if err != nil {
		return microerror.Mask(err)
	}

	hasCilium, err := key.HasCilium(&cr)
	if err != nil {
		return microerror.Mask(err)
	}

	if !hasCilium {
		r.logger.Debugf(ctx, "This cluster has no Cilium.")
		r.logger.Debugf(ctx, "canceling resource")

		return nil
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

	wcCtrlClient := cc.Client.TenantCluster.K8s.CtrlClient()

	// Ensure aws-node daemonset has zero pods.
	ds := &v1.DaemonSet{}
	err = wcCtrlClient.Get(ctx, client.ObjectKey{Name: dsName, Namespace: dsNamespace}, ds)
	if apierrors.IsNotFound(err) {
		// All good.
		r.logger.Debugf(ctx, "Daemonset %q was not found in namespace %q", dsName, dsNamespace)
	} else if err != nil {
		return microerror.Mask(err)
	} else {
		if ds.Status.DesiredNumberScheduled > 0 {
			r.logger.Debugf(ctx, "Daemonset %q/%q still has %d replicas", dsNamespace, dsName, ds.Status.DesiredNumberScheduled)
			r.logger.Debugf(ctx, "canceling resource")

			return nil
		}

		r.logger.Debugf(ctx, "Daemonset %q/%q has no replicas, deleting all resources", dsNamespace, dsName)
	}

	// Get Cluster CR
	cluster := apiv1beta1.Cluster{}
	err = r.ctrlClient.Get(ctx, client.ObjectKey{Namespace: cr.Namespace, Name: cr.Name}, &cluster)
	if err != nil {
		return microerror.Mask(err)
	}

	for _, objToBeDel := range r.objectsToBeDeleted {
		obj := objToBeDel()
		err = wcCtrlClient.Delete(ctx, obj)
		if apierrors.IsNotFound(err) {
			// All good that's what we want.
			continue
		} else if err != nil {
			return microerror.Mask(err)
		}

		name := obj.GetName()
		if obj.GetNamespace() != "" {
			name = fmt.Sprintf("%s/%s", obj.GetNamespace(), name)
		}
		r.logger.Debugf(ctx, "Deleted %s %s", obj.GetObjectKind().GroupVersionKind().Kind, name)
	}

	// Ensure the cilium app has kube proxy enabled.
	if key.ForceDisableCiliumKubeProxyReplacement(cluster) {
		// Ensure no kube-proxy pods are still running.
		{
			r.logger.Debugf(ctx, "Ensuring no kube-proxy pods are still running")

			o := func() error {
				pods := corev1.PodList{}
				err = wcCtrlClient.List(ctx, &pods, client.MatchingLabels{"k8s-app": "kube-proxy"}, client.InNamespace("kube-system"))
				if err != nil {
					return microerror.Mask(err)
				}

				for _, pod := range pods.Items {
					if pod.DeletionTimestamp == nil {
						return microerror.Maskf(kubeProxyStillRunningError, "Kube-proxy pod %s is still running", pod.Name)
					}
				}

				return nil
			}

			b := backoff.NewExponential(30*time.Second, 5*time.Second)
			n := backoff.NewNotifier(r.logger, context.Background())

			err := backoff.RetryNotify(o, b, n)
			if err != nil {
				return microerror.Mask(err)
			}

			r.logger.Debugf(ctx, "Ensured no kube-proxy pods are still running")
		}

		// Remove annotation
		delete(cluster.Annotations, annotation.CiliumForceDisableKubeProxyAnnotation)
		err = r.ctrlClient.Update(ctx, &cluster)
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.Debugf(ctx, "Removed %s annotation from Cluster CR %s", annotation.CiliumForceDisableKubeProxyAnnotation, cluster.Name)
		r.logger.Debugf(ctx, "canceling resource")
		return nil
	}

	if key.CiliumPodsCIDRBlock(cluster) != "" {
		r.logger.Debugf(ctx, "Migrating AWS CNI pod cidr from AWSCluster.Spec.Provider.Pods.CIDRBlock to %q annotation", annotation2.LegacyAwsCniPodCidr)
		cr.Annotations[annotation2.LegacyAwsCniPodCidr] = cr.Spec.Provider.Pods.CIDRBlock

		r.logger.Debugf(ctx, "Migrating cilium pod cidr from %q annotation to AWSCluster.Spec.Provider.Pods.CIDRBlock", annotation.CiliumPodCidr)
		// Update pod cidr on AWSCluster CR
		cr.Spec.Provider.Pods.CIDRBlock = key.CiliumPodsCIDRBlock(cluster)
		err = r.ctrlClient.Update(ctx, &cr)
		if err != nil {
			return microerror.Mask(err)
		}

		// Delete cilium pod cidr annotation from Cluster CR.
		annotations := cluster.Annotations
		delete(annotations, annotation.CiliumPodCidr)
		cluster.Annotations = annotations
		err = r.ctrlClient.Update(ctx, &cluster)
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.Debugf(ctx, "Migrated cilium pod cidr from %q annotation to AWSCluster.Spec.Provider.Pods.CIDRBlock", annotation.CiliumPodCidr)
	}

	return nil
}
