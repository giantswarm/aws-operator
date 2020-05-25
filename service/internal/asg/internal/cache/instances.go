package cache

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/giantswarm/operatorkit/controller/context/cachekeycontext"
	gocache "github.com/patrickmn/go-cache"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/aws-operator/service/controller/key"
)

type Instances struct {
	cache *gocache.Cache
}

func NewInstances() *Instances {
	r := &Instances{
		cache: gocache.New(expiration, expiration/2),
	}

	return r
}

func (r *Instances) Get(ctx context.Context, key string) ([]*ec2.Instance, bool) {
	val, ok := r.cache.Get(key)
	if ok {
		return val.([]*ec2.Instance), true
	}

	return nil, false
}

func (r *Instances) Key(ctx context.Context, obj metav1.Object) string {
	ck, ok := cachekeycontext.FromContext(ctx)
	if ok {
		return fmt.Sprintf("%s/%s", ck, key.ClusterID(obj))
	}

	return ""
}

func (r *Instances) Set(ctx context.Context, key string, val []*ec2.Instance) {
	r.cache.SetDefault(key, val)
}
