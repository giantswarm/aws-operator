package accountid

import (
	"context"
	"fmt"

	"github.com/giantswarm/apiextensions/pkg/apis/infrastructure/v1alpha2"
	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/aws-operator/pkg/label"
	"github.com/giantswarm/aws-operator/service/controller/controllercontext"
)

const (
	Name = "release"
)

type Config struct {
	G8sClient versioned.Interface
	Logger    micrologger.Logger
}

type Resource struct {
	g8sClient versioned.Interface
	logger    micrologger.Logger
}

func New(config Config) (*Resource, error) {
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	r := &Resource{
		logger: config.Logger,
	}

	return r, nil
}

func (r *Resource) Name() string {
	return Name
}

func (r *Resource) addReleaseToContext(ctx context.Context, cr v1alpha2.AWSCluster) error {
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	// Here we take the STS client scoped to the control plane AWS account to
	// lookup its ID. The ID is then set to the controller context.
	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "found the release corresponding to the tenant cluster release label")

		releaseVersion := cr.Labels[label.Release]
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
