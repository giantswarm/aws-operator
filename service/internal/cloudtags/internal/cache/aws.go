package cache

import (
	"context"
	"fmt"

	"github.com/giantswarm/operatorkit/v4/pkg/controller/context/cachekeycontext"
	gocache "github.com/patrickmn/go-cache"
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

func (r *AWS) Get(ctx context.Context, key string) (map[string]string, bool) {
	val, ok := r.cache.Get(key)
	if ok {
		return val.(map[string]string), true
	}

	return nil, false
}

func (r *AWS) Key(ctx context.Context, clusterID string) string {
	ck, ok := cachekeycontext.FromContext(ctx)
	if ok {
		return fmt.Sprintf("%s/%s", ck, clusterID)
	}

	return ""
}

func (r *AWS) Set(ctx context.Context, key string, val map[string]string) {
	r.cache.SetDefault(key, val)
}
