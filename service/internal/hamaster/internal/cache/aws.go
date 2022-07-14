package cache

import (
	"context"
	"fmt"

	infrastructurev1alpha3 "github.com/giantswarm/apiextensions/v6/pkg/apis/infrastructure/v1alpha3"
	"github.com/giantswarm/operatorkit/v7/pkg/controller/context/cachekeycontext"
	gocache "github.com/patrickmn/go-cache"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/aws-operator/v2/service/controller/key"
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

func (r *AWS) Get(ctx context.Context, key string) (infrastructurev1alpha3.AWSControlPlane, bool) {
	val, ok := r.cache.Get(key)
	if ok {
		return val.(infrastructurev1alpha3.AWSControlPlane), true
	}

	return infrastructurev1alpha3.AWSControlPlane{}, false
}

func (r *AWS) Key(ctx context.Context, obj metav1.Object) string {
	ck, ok := cachekeycontext.FromContext(ctx)
	if ok {
		return fmt.Sprintf("%s/%s", ck, key.ClusterID(obj))
	}

	return ""
}

func (r *AWS) Set(ctx context.Context, key string, val infrastructurev1alpha3.AWSControlPlane) {
	r.cache.SetDefault(key, val)
}
