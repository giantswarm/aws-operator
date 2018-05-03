package versionbundle

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"sort"
	"sync"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/go-resty/resty"
	"golang.org/x/sync/errgroup"
)

type CollectorConfig struct {
	FilterFunc func(Bundle) bool
	Logger     micrologger.Logger
	RestClient *resty.Client

	Endpoints []*url.URL
}

type Collector struct {
	filterFunc func(Bundle) bool
	logger     micrologger.Logger
	restClient *resty.Client

	bundles []Bundle
	mutex   sync.Mutex

	endpoints []*url.URL
}

func NewCollector(config CollectorConfig) (*Collector, error) {
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}
	if config.RestClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.RestClient must not be empty", config)
	}

	if len(config.Endpoints) == 0 {
		return nil, microerror.Maskf(invalidConfigError, "%T.Endpoints must not be empty", config)
	}

	c := &Collector{
		filterFunc: config.FilterFunc, // Not required and therefore not validated above.
		logger:     config.Logger,
		restClient: config.RestClient,

		bundles: nil,
		mutex:   sync.Mutex{},

		endpoints: config.Endpoints,
	}

	return c, nil
}

func (c *Collector) Bundles() []Bundle {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	return CopyBundles(c.bundles)
}

type CollectorEndpointResponse struct {
	VersionBundles []Bundle `json:"version_bundles"`
}

func (c *Collector) Collect(ctx context.Context) error {
	c.logger.Log("level", "debug", "message", "collector starts collecting version bundles from endpoints")

	responses := map[string][]byte{}
	{
		var g errgroup.Group

		for _, endpoint := range c.endpoints {
			e := endpoint

			g.Go(func() error {
				c.logger.Log("endpoint", e.String(), "level", "debug", "message", "collector requesting version bundles from endpoint")

				res, err := c.restClient.NewRequest().Get(e.String())
				if err != nil {
					return microerror.Mask(err)
				}

				c.logger.Log("endpoint", e.String(), "level", "debug", "message", "collector received version bundles from endpoint")

				c.mutex.Lock()
				responses[e.String()] = res.Body()
				c.mutex.Unlock()

				return nil
			})

			err := g.Wait()
			if err != nil {
				return microerror.Mask(err)
			}
		}
	}

	var bundles []Bundle
	{

		for e, b := range responses {
			var r CollectorEndpointResponse
			err := json.Unmarshal(b, &r)
			if err != nil {
				return microerror.Mask(err)
			}

			var filteredBundles []Bundle

			if c.filterFunc != nil {
				for _, b := range r.VersionBundles {
					if c.filterFunc(b) {
						filteredBundles = append(filteredBundles, b)
					} else {
						c.logger.Log("endpoint", e, "level", "debug", "message", fmt.Sprintf("filterFunc rejected: %#v", b))
					}

				}
			} else {
				filteredBundles = r.VersionBundles
			}

			c.logger.Log("endpoint", e, "level", "debug", "message", fmt.Sprintf("collector found %d version bundles from endpoint. %d filtered out.", len(r.VersionBundles), (len(r.VersionBundles)-len(filteredBundles))))
			bundles = append(bundles, filteredBundles...)
		}
	}

	sort.Sort(SortBundlesByVersion(bundles))
	sort.Stable(SortBundlesByName(bundles))

	{
		c.mutex.Lock()
		c.bundles = bundles
		c.mutex.Unlock()
	}

	c.logger.Log("level", "debug", "message", "collector finishes collecting version bundles from endpoints")

	return nil
}
