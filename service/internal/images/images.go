package images

import (
	"context"

	infrastructurev1alpha2 "github.com/giantswarm/apiextensions/pkg/apis/infrastructure/v1alpha2"
	releasev1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/release/v1alpha1"
	"github.com/giantswarm/k8sclient/v3/pkg/k8sclient"
	k8scloudconfig "github.com/giantswarm/k8scloudconfig/v6/pkg/template"
	"github.com/giantswarm/microerror"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/aws-operator/pkg/label"
	"github.com/giantswarm/aws-operator/service/controller/key"
	"github.com/giantswarm/aws-operator/service/internal/images/internal/cache"
)

type Config struct {
	K8sClient k8sclient.Interface

	RegistryDomain string
}

type Images struct {
	k8sClient k8sclient.Interface

	clusterCache *cache.Cluster
	releaseCache *cache.Release

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

		clusterCache: cache.NewCluster(),
		releaseCache: cache.NewRelease(),

		registryDomain: c.RegistryDomain,
	}

	return i, nil
}

func (i *Images) AMI(ctx context.Context, obj interface{}) (string, error) {
	cr, err := meta.Accessor(obj)
	if err != nil {
		return "", microerror.Mask(err)
	}

	cl, err := i.cachedCluster(ctx, cr)
	if err != nil {
		return "", microerror.Mask(err)
	}

	re, err := i.cachedRelease(ctx, cr)
	if err != nil {
		return "", microerror.Mask(err)
	}

	ami, err := key.AMI(key.Region(cl), re)
	if err != nil {
		return "", microerror.Mask(err)
	}

	return ami, nil
}

func (i *Images) CC(ctx context.Context, obj interface{}) (k8scloudconfig.Images, error) {
	cr, err := meta.Accessor(obj)
	if err != nil {
		return k8scloudconfig.Images{}, microerror.Mask(err)
	}

	re, err := i.cachedRelease(ctx, cr)
	if err != nil {
		return k8scloudconfig.Images{}, microerror.Mask(err)
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

func (i *Images) cachedCluster(ctx context.Context, cr metav1.Object) (infrastructurev1alpha2.AWSCluster, error) {
	var err error
	var ok bool

	var cluster infrastructurev1alpha2.AWSCluster
	{
		ck := i.clusterCache.Key(ctx, cr)

		if ck == "" {
			cluster, err = i.lookupCluster(ctx, cr)
			if err != nil {
				return infrastructurev1alpha2.AWSCluster{}, microerror.Mask(err)
			}
		} else {
			cluster, ok = i.clusterCache.Get(ctx, ck)
			if !ok {
				cluster, err = i.lookupCluster(ctx, cr)
				if err != nil {
					return infrastructurev1alpha2.AWSCluster{}, microerror.Mask(err)
				}

				i.clusterCache.Set(ctx, ck, cluster)
			}
		}
	}

	return cluster, nil
}

func (i *Images) cachedRelease(ctx context.Context, cr metav1.Object) (releasev1alpha1.Release, error) {
	var err error
	var ok bool

	var re releasev1alpha1.Release
	{
		ck := i.releaseCache.Key(ctx, cr)

		if ck == "" {
			re, err = i.lookupRelease(ctx, cr)
			if err != nil {
				return releasev1alpha1.Release{}, microerror.Mask(err)
			}
		} else {
			re, ok = i.releaseCache.Get(ctx, ck)
			if !ok {
				re, err = i.lookupRelease(ctx, cr)
				if err != nil {
					return releasev1alpha1.Release{}, microerror.Mask(err)
				}

				i.releaseCache.Set(ctx, ck, re)
			}
		}
	}

	return re, nil
}

func (i *Images) lookupCluster(ctx context.Context, cr metav1.Object) (infrastructurev1alpha2.AWSCluster, error) {
	var list infrastructurev1alpha2.AWSClusterList

	err := i.k8sClient.CtrlClient().List(
		ctx,
		&list,
		client.InNamespace(cr.GetNamespace()),
		client.MatchingLabels{label.Cluster: key.ClusterID(cr)},
	)
	if err != nil {
		return infrastructurev1alpha2.AWSCluster{}, microerror.Mask(err)
	}

	if len(list.Items) == 0 {
		return infrastructurev1alpha2.AWSCluster{}, microerror.Mask(notFoundError)
	}
	if len(list.Items) > 1 {
		return infrastructurev1alpha2.AWSCluster{}, microerror.Mask(tooManyCRsError)
	}

	return list.Items[0], nil
}

func (i *Images) lookupRelease(ctx context.Context, cr metav1.Object) (releasev1alpha1.Release, error) {
	var re releasev1alpha1.Release

	err := i.k8sClient.CtrlClient().Get(
		ctx,
		types.NamespacedName{Name: key.ReleaseName(key.ReleaseVersion(cr))},
		&re,
	)
	if err != nil {
		return releasev1alpha1.Release{}, microerror.Mask(err)
	}

	return re, nil
}
