package healthz

import (
	"context"
	"net/http"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	kitendpoint "github.com/go-kit/kit/endpoint"
	kithttp "github.com/go-kit/kit/transport/http"

	"github.com/giantswarm/aws-operator/server/middleware"
	"github.com/giantswarm/aws-operator/service"
	"github.com/giantswarm/aws-operator/service/healthz"
)

const (
	// Method is the HTTP method this endpoint is registered for.
	Method = "GET"
	// Name identifies the endpoint. It is aligned to the package path.
	Name = "healthz"
	// Path is the HTTP request path this endpoint is registered for.
	Path = "/healthz"
)

// Config represents the configuration used to create a healthz endpoint.
type Config struct {
	// Dependencies.
	Logger     micrologger.Logger
	Middleware *middleware.Middleware
	Service    *service.Service
}

// DefaultConfig provides a default configuration to create a new healthz
// endpoint by best effort.
func DefaultConfig() Config {
	return Config{
		// Dependencies.
		Logger:     nil,
		Middleware: nil,
		Service:    nil,
	}
}

// New creates a new configured healthz endpoint.
func New(config Config) (*Endpoint, error) {
	// Dependencies.
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "logger must not be empty")
	}
	if config.Middleware == nil {
		return nil, microerror.Maskf(invalidConfigError, "middleware must not be empty")
	}
	if config.Service == nil {
		return nil, microerror.Maskf(invalidConfigError, "service must not be empty")
	}

	newEndpoint := &Endpoint{
		Config: config,
	}

	return newEndpoint, nil
}

type Endpoint struct {
	Config
}

func (e *Endpoint) Decoder() kithttp.DecodeRequestFunc {
	return func(ctx context.Context, r *http.Request) (interface{}, error) {
		return nil, nil
	}
}

func (e *Endpoint) Encoder() kithttp.EncodeResponseFunc {
	return func(ctx context.Context, w http.ResponseWriter, response interface{}) error {
		w.WriteHeader(http.StatusOK)
		return nil
	}
}

func (e *Endpoint) Endpoint() kitendpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		_, err := e.Service.Healthz.Check(ctx, healthz.DefaultRequest())
		if err != nil {
			return nil, microerror.Mask(err)
		}

		return nil, nil
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
