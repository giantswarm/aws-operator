package cache

import (
	"context"
	"fmt"

	infrastructurev1alpha3 "github.com/giantswarm/apiextensions/v6/pkg/apis/infrastructure/v1alpha3"
	"github.com/giantswarm/operatorkit/v7/pkg/controller/context/cachekeycontext"
	gocache "github.com/patrickmn/go-cache"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/aws-operator/v13/service/controller/key"
)

type G8s struct {
	cache *gocache.Cache
}

func NewG8s() *G8s {
	r := &G8s{
		cache: gocache.New(expiration, expiration/2),
	}

	return r
}

func (r *G8s) Get(ctx context.Context, key string) (infrastructurev1alpha3.G8sControlPlane, bool) {
	val, ok := r.cache.Get(key)
	if ok {
		return val.(infrastructurev1alpha3.G8sControlPlane), true
	}

	return infrastructurev1alpha3.G8sControlPlane{}, false
}

func (r *G8s) Key(ctx context.Context, obj metav1.Object) string {
	ck, ok := cachekeycontext.FromContext(ctx)
	if ok {
		return fmt.Sprintf("%s/%s", ck, key.ClusterID(obj))
	}

	return ""
}

func (r *G8s) Set(ctx context.Context, key string, val infrastructurev1alpha3.G8sControlPlane) {
	r.cache.SetDefault(key, val)
}
