package healthz

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	kitendpoint "github.com/go-kit/kit/endpoint"
	kithttp "github.com/go-kit/kit/transport/http"

	"github.com/giantswarm/microendpoint/service/healthz"
)

const (
	// Method is the HTTP method this endpoint is register for.
	Method = "GET"
	// Name identifies the endpoint. It is aligned to the package path.
	Name = "healthz"
	// Path is the HTTP request path this endpoint is registered for.
	Path = "/healthz"
)

// Config represents the configured used to create a healthz endpoint.
type Config struct {
	// Dependencies.
	Logger   micrologger.Logger
	Services []healthz.Service
}

// DefaultConfig provides a default configuration to create a new healthz
// endpoint by best effort.
func DefaultConfig() Config {
	return Config{
		// Dependencies.
		Logger:   nil,
		Services: nil,
	}
}

// New creates a new configured healthz endpoint.
func New(config Config) (*Endpoint, error) {
	// Dependencies.
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.Logger must not be empty")
	}
	if len(config.Services) == 0 {
		c := healthz.Config{
			Logger: config.Logger,
		}

		h, err := healthz.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		config.Services = append(config.Services, h)
	}

	newEndpoint := &Endpoint{
		logger:   config.Logger,
		services: config.Services,
	}

	return newEndpoint, nil
}

type Endpoint struct {
	// Dependencies.
	logger   micrologger.Logger
	services []healthz.Service
}

func (e *Endpoint) Decoder() kithttp.DecodeRequestFunc {
	return func(ctx context.Context, r *http.Request) (interface{}, error) {
		return nil, nil
	}
}

func (e *Endpoint) Encoder() kithttp.EncodeResponseFunc {
	return func(ctx context.Context, w http.ResponseWriter, response interface{}) error {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")

		rs, ok := response.([]healthz.Response)
		if !ok {
			return microerror.Maskf(wrongTypeError, "expected '%T' got '%T'", []healthz.Response{}, response)
		}
		if healthz.Responses(rs).HasFailed() {
			for _, r := range rs {
				if r.Failed {
					e.logger.Log("error", "health check failed", "healthCheckDescription", r.Description, "healthCheckMessage", r.Message)
				}
			}

			w.WriteHeader(http.StatusInternalServerError)
		}

		return json.NewEncoder(w).Encode(response)
	}
}

func (e *Endpoint) Endpoint() kitendpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		var healthzResponses []healthz.Response

		for _, s := range e.services {
			res, err := s.GetHealthz(ctx)
			if err != nil {
				return nil, microerror.Mask(err)
			}

			healthzResponses = append(healthzResponses, res)
		}

		return healthzResponses, nil
	}
}

func (e *Endpoint) Method() string {
	return Method
}

func (e *Endpoint) Middlewares() []kitendpoint.Middleware {
	return []kitendpoint.Middleware{}
}

func (e *Endpoint) Name() string {
	return Name
}

func (e *Endpoint) Path() string {
	return Path
}
