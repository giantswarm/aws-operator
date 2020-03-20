package cache

import (
	"time"

	gocache "github.com/patrickmn/go-cache"
)

type Float64Cache struct {
	underlying *gocache.Cache
}

func NewFloat64Cache(expiration time.Duration) *Float64Cache {
	c := &Float64Cache{
		// Clean up period is set to half of the expiration, which means values are
		// checked to be cleaned at least once before the expiration time.
		underlying: gocache.New(expiration, expiration/2),
	}

	return c
}

func (c *Float64Cache) Get(k string) (float64, bool) {
	v, _ := c.underlying.Get(k)
	if v == nil {
		return 0, false
	}

	vn, ok := v.(float64)
	if !ok {
		return 0, false
	}

	return vn, true
}

func (c *Float64Cache) Set(k string, v float64) {
	c.underlying.Set(k, v, 0)
}
