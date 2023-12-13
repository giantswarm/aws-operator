package kms

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/giantswarm/operatorkit/v8/pkg/controller/context/cachekeycontext"
	gocache "github.com/patrickmn/go-cache"
)

const (
	expiration = 5 * time.Minute
)

type Cache struct {
	cache *gocache.Cache
}

func NewCache() *Cache {
	r := &Cache{
		cache: gocache.New(expiration, expiration/2),
	}

	return r
}

func (r *Cache) Get(ctx context.Context, key string) (*kms.DescribeKeyOutput, bool) {
	val, ok := r.cache.Get(key)
	if ok {
		return val.(*kms.DescribeKeyOutput), true
	}

	return &kms.DescribeKeyOutput{}, false
}

func (r *Cache) Key(ctx context.Context, id string) string {
	ck, ok := cachekeycontext.FromContext(ctx)
	if ok {
		return fmt.Sprintf("%s/%s", ck, id)
	}

	return ""
}

func (r *Cache) Set(ctx context.Context, key string, val *kms.DescribeKeyOutput) {
	r.cache.SetDefault(key, val)
}
