package cache

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/giantswarm/operatorkit/controller/context/cachekeycontext"
	gocache "github.com/patrickmn/go-cache"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/aws-operator/service/controller/key"
)

type ASGs struct {
	cache *gocache.Cache
}

func NewASGs() *ASGs {
	a := &ASGs{
		cache: gocache.New(expiration, expiration/2),
	}

	return a
}

func (a *ASGs) Get(ctx context.Context, key string) ([]*autoscaling.Group, bool) {
	val, ok := a.cache.Get(key)
	if ok {
		return val.([]*autoscaling.Group), true
	}

	return nil, false
}

func (a *ASGs) Key(ctx context.Context, obj metav1.Object) string {
	ck, ok := cachekeycontext.FromContext(ctx)
	if ok {
		return fmt.Sprintf("%s/%s", ck, key.ClusterID(obj))
	}

	return ""
}

func (a *ASGs) Set(ctx context.Context, key string, val []*autoscaling.Group) {
	a.cache.SetDefault(key, val)
}
