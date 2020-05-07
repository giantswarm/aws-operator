package images

import (
	"context"

	"github.com/giantswarm/apiextensions/pkg/apis/release/v1alpha1"
	"github.com/giantswarm/k8sclient"
	k8scloudconfig "github.com/giantswarm/k8scloudconfig/v6/v_6_0_0"
	"github.com/giantswarm/microerror"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/types"

	"github.com/giantswarm/aws-operator/service/controller/key"
)

type Config struct {
	K8sClient k8sclient.Interface

	RegistryDomain string
}

type Images struct {
	k8sClient k8sclient.Interface

	registryDomain string
}

func New(c Config) (*Images, error) {
	if c.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.K8sClient must not be empty", c)
	}

	if c.RegistryDomain == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.RegistryDomain must not be empty", c)
	}

	i := &Images{
		k8sClient: c.K8sClient,

		registryDomain: c.RegistryDomain,
	}

	return i, nil
}

func (i *Images) ForRelease(ctx context.Context, obj interface{}) (k8scloudconfig.Images, error) {
	cr, err := meta.Accessor(obj)
	if err != nil {
		return k8scloudconfig.Images{}, microerror.Mask(err)
	}

	var re v1alpha1.Release
	{
		err := i.k8sClient.CtrlClient().Get(
			ctx,
			types.NamespacedName{Name: key.ReleaseName(key.ReleaseVersion(cr))},
			&re,
		)
		if err != nil {
			return k8scloudconfig.Images{}, microerror.Mask(err)
		}
	}

	var im k8scloudconfig.Images
	{
		v, err := k8scloudconfig.ExtractComponentVersions(re.Spec.Components)
		if err != nil {
			return k8scloudconfig.Images{}, microerror.Mask(err)
		}

		v.Kubectl = key.KubectlVersion
		v.KubernetesAPIHealthz = key.KubernetesAPIHealthzVersion
		v.KubernetesNetworkSetupDocker = key.K8sSetupNetworkEnvironment

		im = k8scloudconfig.BuildImages(i.registryDomain, v)
	}

	return im, nil
}