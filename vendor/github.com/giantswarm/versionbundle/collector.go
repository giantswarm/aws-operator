package versionbundle

import (
	"context"
	"encoding/json"
	"net/url"
	"sync"

	"github.com/giantswarm/microerror"
	"github.com/go-resty/resty"
	"golang.org/x/sync/errgroup"
)

type CollectorConfig struct {
	RestClient *resty.Client

	Endpoints []*url.URL
}

func DefaultCollectorConfig() CollectorConfig {
	return CollectorConfig{
		RestClient: nil,

		Endpoints: nil,
	}
}

type Collector struct {
	restClient *resty.Client

	bundles []Bundle
	mutex   sync.Mutex

	endpoints []*url.URL
}

func NewCollector(config CollectorConfig) (*Collector, error) {
	if config.RestClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.RestClient must not be empty")
	}

	if len(config.Endpoints) == 0 {
		return nil, microerror.Maskf(invalidConfigError, "config.Endpoints must not be empty")
	}

	c := &Collector{
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
	var responses [][]byte
	{
		var g errgroup.Group
		responses = make([][]byte, len(c.endpoints))

		for i, e := range c.endpoints {
			i, e := i, e

			g.Go(func() error {
				res, err := c.restClient.NewRequest().Get(e.String())
				if err != nil {
					return microerror.Mask(err)
				}

				responses[i] = res.Body()

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

		for _, b := range responses {
			var er CollectorEndpointResponse
			err := json.Unmarshal(b, &er)
			if err != nil {
				return microerror.Mask(err)
			}

			for _, bundle := range er.VersionBundles {
				bundles = append(bundles, bundle)
			}
		}
	}

	{
		c.mutex.Lock()
		c.bundles = bundles
		c.mutex.Unlock()
	}

	return nil
}
