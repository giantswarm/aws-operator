package cache

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/giantswarm/operatorkit/v7/pkg/controller/context/cachekeycontext"
	gocache "github.com/patrickmn/go-cache"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/aws-operator/v13/service/controller/key"
)

type Instances struct {
	cache *gocache.Cache
}

func NewInstances() *Instances {
	i := &Instances{
		cache: gocache.New(expiration, expiration/2),
	}

	return i
}

func (i *Instances) Get(ctx context.Context, key string) ([]*ec2.Instance, bool) {
	val, ok := i.cache.Get(key)
	if ok {
		return val.([]*ec2.Instance), true
	}

	return nil, false
}

func (i *Instances) Key(ctx context.Context, obj metav1.Object) string {
	ck, ok := cachekeycontext.FromContext(ctx)
	if ok {
		return fmt.Sprintf("%s/%s", ck, key.ClusterID(obj))
	}

	return ""
}

func (i *Instances) Set(ctx context.Context, key string, val []*ec2.Instance) {
	i.cache.SetDefault(key, val)
}
