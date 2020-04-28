package release

import (
	"context"
	"fmt"

	"github.com/giantswarm/microerror"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/aws-operator/pkg/label"
	"github.com/giantswarm/aws-operator/service/controller/controllercontext"
	"github.com/giantswarm/aws-operator/service/controller/key"
)

func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	cr, err := key.ToCustomObject(obj)
	if err != nil {
		return microerror.Mask(err)
	}
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "found the release corresponding to the tenant cluster release label")

		releaseVersion := cr.Labels[label.ReleaseVersion]
		releaseName := fmt.Sprintf("v%s", releaseVersion)
		release, err := r.g8sClient.ReleaseV1alpha1().Releases().Get(releaseName, metav1.GetOptions{})
		if err != nil {
			return microerror.Mask(err)
		}

		cc.Spec.TenantCluster.Release = *release

		r.logger.LogCtx(ctx, "level", "debug", "message", "found the release corresponding to the tenant cluster release label")
	}

	return nil
}
