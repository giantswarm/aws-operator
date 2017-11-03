package versionbundle

import (
	"net/url"

	"github.com/giantswarm/microerror"
	"github.com/go-resty/resty"
)

type CollectorConfig struct {
	RestClient *resty.Client

	Endpoints []url.URL
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

	endpoints []url.URL
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

		endpoints: config.Endpoints,
	}

	return c, nil
}

func (c *Collector) Bundles() []Bundle {
	return c.bundles
}

// TODO
func (c *Collector) Collect() error {
	return nil
}
