package cache

import (
	"context"
	"fmt"

	infrastructurev1alpha2 "github.com/giantswarm/apiextensions/v3/pkg/apis/infrastructure/v1alpha2"
	"github.com/giantswarm/operatorkit/v5/pkg/controller/context/cachekeycontext"
	gocache "github.com/patrickmn/go-cache"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/aws-operator/service/controller/key"
)

type AWS struct {
	cache *gocache.Cache
}

func NewAWS() *AWS {
	r := &AWS{
		cache: gocache.New(expiration, expiration/2),
	}

	return r
}

func (r *AWS) Get(ctx context.Context, key string) (infrastructurev1alpha2.AWSControlPlane, bool) {
	val, ok := r.cache.Get(key)
	if ok {
		return val.(infrastructurev1alpha2.AWSControlPlane), true
	}

	return infrastructurev1alpha2.AWSControlPlane{}, false
}

func (r *AWS) Key(ctx context.Context, obj metav1.Object) string {
	ck, ok := cachekeycontext.FromContext(ctx)
	if ok {
		return fmt.Sprintf("%s/%s", ck, key.ClusterID(obj))
	}

	return ""
}

func (r *AWS) Set(ctx context.Context, key string, val infrastructurev1alpha2.AWSControlPlane) {
	r.cache.SetDefault(key, val)
}
