package endpoint

import (
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"

	"github.com/giantswarm/aws-operator/server/endpoint/healthz"
	"github.com/giantswarm/aws-operator/server/endpoint/version"
	"github.com/giantswarm/aws-operator/server/middleware"
	"github.com/giantswarm/aws-operator/service"
)

// Config represents the configuration used to create a endpoint.
type Config struct {
	// Dependencies.
	Logger     micrologger.Logger
	Middleware *middleware.Middleware
	Service    *service.Service
}

// DefaultConfig provides a default configuration to create a new endpoint by
// best effort.
func DefaultConfig() Config {
	return Config{
		// Dependencies.
		Logger:     nil,
		Middleware: nil,
		Service:    nil,
	}
}

// New creates a new configured endpoint.
func New(config Config) (*Endpoint, error) {
	var err error

	var healthzEndpoint *healthz.Endpoint
	{
		healthzConfig := healthz.DefaultConfig()
		healthzConfig.Logger = config.Logger
		healthzConfig.Middleware = config.Middleware
		healthzConfig.Service = config.Service
		healthzEndpoint, err = healthz.New(healthzConfig)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var versionEndpoint *version.Endpoint
	{
		versionConfig := version.DefaultConfig()
		versionConfig.Logger = config.Logger
		versionConfig.Middleware = config.Middleware
		versionConfig.Service = config.Service
		versionEndpoint, err = version.New(versionConfig)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	newEndpoint := &Endpoint{
		Healthz: healthzEndpoint,
		Version: versionEndpoint,
	}

	return newEndpoint, nil
}

// Endpoint is the endpoint collection.
type Endpoint struct {
	Healthz *healthz.Endpoint
	Version *version.Endpoint
}
