package endpoint

import (
	"github.com/giantswarm/microendpoint/endpoint/healthz"
	"github.com/giantswarm/microendpoint/endpoint/version"
	healthzservice "github.com/giantswarm/microendpoint/service/healthz"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"

	"github.com/giantswarm/kvm-operator/server/middleware"
	"github.com/giantswarm/kvm-operator/service"
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
		healthzConfig.Services = []healthzservice.Service{
			config.Service.Healthz.K8s,
		}
		healthzEndpoint, err = healthz.New(healthzConfig)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var versionEndpoint *version.Endpoint
	{
		versionConfig := version.DefaultConfig()
		versionConfig.Logger = config.Logger
		versionConfig.Service = config.Service.Version
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
