package cache

import (
	"context"

	releasev1alpha1 "github.com/giantswarm/apiextensions/v3/pkg/apis/release/v1alpha1"
	"github.com/giantswarm/operatorkit/v5/pkg/controller/context/cachekeycontext"
	gocache "github.com/patrickmn/go-cache"
)

type Release struct {
	cache *gocache.Cache
}

func NewRelease() *Release {
	r := &Release{
		cache: gocache.New(expiration, expiration/2),
	}

	return r
}

func (r *Release) Get(ctx context.Context, key string) (releasev1alpha1.Release, bool) {
	val, ok := r.cache.Get(key)
	if ok {
		return val.(releasev1alpha1.Release), true
	}

	return releasev1alpha1.Release{}, false
}

func (r *Release) Key(ctx context.Context, version string) string {
	_, ok := cachekeycontext.FromContext(ctx)
	if ok {
		return version
	}

	return ""
}

func (r *Release) Set(ctx context.Context, key string, val releasev1alpha1.Release) {
	r.cache.SetDefault(key, val)
}
