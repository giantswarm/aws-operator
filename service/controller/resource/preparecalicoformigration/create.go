package preparecalicoformigration

import (
	"context"
	"fmt"
	"strings"

	"github.com/blang/semver"
	"github.com/giantswarm/microerror"
	v1 "k8s.io/api/apps/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	apiv1beta1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/aws-operator/v14/service/controller/controllercontext"
	"github.com/giantswarm/aws-operator/v14/service/controller/key"
)

const (
	dsNamespace    = "kube-system"
	dsName         = "calico-node"
	desiredVersion = "v3.22.3"
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

	// Only run this if cilium ENI ipam mode is enabled.
	cluster := apiv1beta1.Cluster{}
	err = r.ctrlClient.Get(ctx, client.ObjectKey{Namespace: cr.Namespace, Name: key.ClusterID(&cr)}, &cluster)
	if err != nil {
		return microerror.Mask(err)
	}

	if !key.IsCiliumEniModeEnabled(cluster) {
		r.logger.Debugf(ctx, "Cilium ipam mode is not 'eni', nothing to do.")
		r.logger.Debugf(ctx, "canceling resource")
		return nil
	}

	wcCtrlClient := cc.Client.TenantCluster.K8s.CtrlClient()

	ds := &v1.DaemonSet{}
	err = wcCtrlClient.Get(ctx, client.ObjectKey{Name: dsName, Namespace: dsNamespace}, ds)
	if apierrors.IsNotFound(err) {
		// All good.
		r.logger.Debugf(ctx, "Daemonset %q was not found in namespace %q, nothing to do", dsName, dsNamespace)
		return nil
	} else if err != nil {
		return microerror.Mask(err)
	}

	// Ensure calico daemonset has image new enough.
	currentImage := ds.Spec.Template.Spec.Containers[0].Image
	desiredImage, err := ensureImageVersionIsUpToDate(currentImage)
	if err != nil {
		r.logger.Debugf(ctx, "Error ensuring calico image version for image %q", currentImage)
		return microerror.Mask(err)
	}

	if currentImage != desiredImage {
		r.logger.Debugf(ctx, "Calico image needs updating from %q to %q", currentImage, desiredImage)
		ds.Spec.Template.Spec.Containers[0].Image = desiredImage

		err = wcCtrlClient.Update(ctx, ds)
		if err != nil {
			return microerror.Mask(err)
		}

		// Wait for next reconciliation loop.
		r.logger.Debugf(ctx, "Daemonset %q was updated", dsName)
	} else {
		r.logger.Debugf(ctx, "Calico image looks good")
	}

	return nil
}

func ensureImageVersionIsUpToDate(image string) (string, error) {
	// parse the image which is in this format
	// docker.io/giantswarm/node:v3.21.5
	// and ensure the image version is at least `minDesiredVersion`

	versionStr := strings.Split(image, ":")[1]

	version, err := semver.ParseTolerant(versionStr)
	if err != nil {
		return "", microerror.Mask(err)
	}

	desired, err := semver.ParseTolerant(desiredVersion)
	if err != nil {
		return "", microerror.Mask(err)
	}

	if version.LT(desired) {
		return fmt.Sprintf("%s:%s", strings.Split(image, ":")[0], desiredVersion), nil
	}

	return image, nil
}
