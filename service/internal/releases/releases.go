package releases

import (
	"context"

	"github.com/giantswarm/k8sclient/v7/pkg/k8sclient"
	"github.com/giantswarm/microerror"
	releasev1alpha1 "github.com/giantswarm/release-operator/v3/api/v1alpha1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"

	"github.com/giantswarm/aws-operator/v2/service/controller/key"
	"github.com/giantswarm/aws-operator/v2/service/internal/releases/internal/cache"
)

type Config struct {
	K8sClient k8sclient.Interface
}

type Releases struct {
	k8sClient k8sclient.Interface

	releaseCache *cache.Release
}

func New(c Config) (*Releases, error) {
	if c.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.K8sClient must not be empty", c)
	}

	r := &Releases{
		k8sClient: c.K8sClient,

		releaseCache: cache.NewRelease(),
	}

	return r, nil
}

func (r *Releases) Release(ctx context.Context, version string) (releasev1alpha1.Release, error) {
	re, err := r.cachedRelease(ctx, version)
	if err != nil {
		return releasev1alpha1.Release{}, microerror.Mask(err)
	}

	return re, nil
}

func (r *Releases) cachedRelease(ctx context.Context, version string) (releasev1alpha1.Release, error) {
	var err error
	var ok bool

	var re releasev1alpha1.Release
	{
		rk := r.releaseCache.Key(ctx, version)

		if rk == "" {
			re, err = r.lookupRelease(ctx, version)
			if err != nil {
				return releasev1alpha1.Release{}, microerror.Mask(err)
			}
		} else {
			re, ok = r.releaseCache.Get(ctx, version)
			if !ok {
				re, err = r.lookupRelease(ctx, version)
				if err != nil {
					return releasev1alpha1.Release{}, microerror.Mask(err)
				}

				r.releaseCache.Set(ctx, version, re)
			}
		}
	}

	return re, nil
}

func (r *Releases) lookupRelease(ctx context.Context, version string) (releasev1alpha1.Release, error) {
	var re releasev1alpha1.Release

	err := r.k8sClient.CtrlClient().Get(
		ctx,
		types.NamespacedName{Name: key.ReleaseName(version)},
		&re,
	)
	if apierrors.IsNotFound(err) {
		return releasev1alpha1.Release{}, microerror.Mask(notFoundError)
	} else if err != nil {
		return releasev1alpha1.Release{}, microerror.Mask(err)
	}

	return re, nil
}
