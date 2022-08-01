package awscnicleaner

import (
	"context"
	"fmt"

	"github.com/giantswarm/microerror"
	v1 "k8s.io/api/apps/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/aws-operator/v13/service/controller/controllercontext"
)

const (
	dsNamespace = "kube-system"
	dsName      = "aws-node"
)

func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	var err error
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	if cc.Client.TenantCluster.K8s == nil {
		r.logger.Debugf(ctx, "kubernetes clients are not available in controller context yet")
		r.logger.Debugf(ctx, "canceling resource")

		return nil
	}

	ctrlClient := cc.Client.TenantCluster.K8s.CtrlClient()

	// Ensure aws-node daemonset has zero pods.
	ds := &v1.DaemonSet{}
	err = ctrlClient.Get(ctx, client.ObjectKey{Name: dsName, Namespace: dsNamespace}, ds)
	if apierrors.IsNotFound(err) {
		// All good.
		r.logger.Debugf(ctx, "Daemonset %q was not found in namespace %q", dsName, dsNamespace)
	} else if err != nil {
		return microerror.Mask(err)
	}

	if ds != nil {
		if ds.Status.DesiredNumberScheduled > 0 {
			r.logger.Debugf(ctx, "Daemonset %q/%q still has %d replicas", dsNamespace, dsName, ds.Status.DesiredNumberScheduled)
			r.logger.Debugf(ctx, "canceling resource")

			return nil
		}
	}

	r.logger.Debugf(ctx, "Daemonset %q/%q has no replicas, deleting all resources", dsNamespace, dsName)

	for _, objToBeDel := range r.objectsToBeDeleted {
		obj := objToBeDel()
		err = ctrlClient.Delete(ctx, obj)
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

	return nil
}
