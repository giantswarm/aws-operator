package endpoint

import (
	"github.com/giantswarm/microendpoint/endpoint/healthz"
	versionendpoint "github.com/giantswarm/microendpoint/endpoint/version"
	healthzservice "github.com/giantswarm/microendpoint/service/healthz"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"

	"github.com/giantswarm/aws-operator/service"
)

type Config struct {
	Logger  micrologger.Logger
	Service *service.Service
}

func New(config Config) (*Endpoint, error) {
	var err error

	var healthzEndpoint *healthz.Endpoint
	{
		c := healthz.Config{
			Logger: config.Logger,
			Services: []healthzservice.Service{
				config.Service.Healthz,
			},
		}

		healthzEndpoint, err = healthz.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var versionEndpoint *versionendpoint.Endpoint
	{
		c := versionendpoint.Config{
			Logger:  config.Logger,
			Service: config.Service.Version,
		}

		versionEndpoint, err = versionendpoint.New(c)
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
	Version *versionendpoint.Endpoint
}
