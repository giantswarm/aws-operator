package awscnicleaner

import (
	"context"
	"fmt"

	"github.com/giantswarm/k8smetadata/pkg/annotation"
	"github.com/giantswarm/microerror"
	v1 "k8s.io/api/apps/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/aws-operator/v13/service/controller/controllercontext"
	"github.com/giantswarm/aws-operator/v13/service/controller/key"
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

	if key.CiliumPodsCIDRBlock(cr) != "" {
		r.logger.Debugf(ctx, "Migrating cilium pod cidr from %q annotation to AWSCluster.Spec.Provider.Pods.CIDRBlock", annotation.CiliumPodCidr)

		cr.Spec.Provider.Pods.CIDRBlock = key.CiliumPodsCIDRBlock(cr)

		annotations := cr.Annotations
		delete(annotations, annotation.CiliumPodCidr)
		cr.Annotations = annotations

		err = r.ctrlClient.Update(ctx, &cr)
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.Debugf(ctx, "Migrated cilium pod cidr from %q annotation to AWSCluster.Spec.Provider.Pods.CIDRBlock", annotation.CiliumPodCidr)
	}

	return nil
}
