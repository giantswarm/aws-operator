package cache

import (
	"context"
	"fmt"

	"github.com/giantswarm/operatorkit/controller/context/cachekeycontext"
	gocache "github.com/patrickmn/go-cache"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apiv1alpha2 "sigs.k8s.io/cluster-api/api/v1alpha2"

	"github.com/giantswarm/aws-operator/service/controller/key"
)

type CAPI struct {
	cache *gocache.Cache
}

func NewCAPI() *CAPI {
	r := &CAPI{
		cache: gocache.New(expiration, expiration/2),
	}

	return r
}

func (r *CAPI) Get(ctx context.Context, key string) (apiv1alpha2.Cluster, bool) {
	val, ok := r.cache.Get(key)
	if ok {
		return val.(apiv1alpha2.Cluster), true
	}

	return apiv1alpha2.Cluster{}, false
}

func (r *CAPI) Key(ctx context.Context, obj metav1.Object) string {
	ck, ok := cachekeycontext.FromContext(ctx)
	if ok {
		return fmt.Sprintf("%s/%s", ck, key.ClusterID(obj))
	}

	return ""
}

func (r *CAPI) Set(ctx context.Context, key string, val apiv1alpha2.Cluster) {
	r.cache.SetDefault(key, val)
}
