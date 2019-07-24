package version

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	kitendpoint "github.com/go-kit/kit/endpoint"
	kithttp "github.com/go-kit/kit/transport/http"

	"github.com/giantswarm/microendpoint/service/version"
)

const (
	// Method is the HTTP method this endpoint is registered for.
	Method = "GET"
	// Name identifies the endpoint. It is aligned to the package path.
	Name = "version"
	// Path is the HTTP request path this endpoint is registered for.
	Path = "/"
)

// Config represents the configuration used to create a version endpoint.
type Config struct {
	// Dependencies.
	Logger  micrologger.Logger
	Service *version.Service
}

// New creates a new configured version endpoint.
func New(config Config) (*Endpoint, error) {
	// Dependencies.
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "logger must not be empty")
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
		w.Header().Set("Content-Type", "application/json; charset=utf-8")

		return json.NewEncoder(w).Encode(response)
	}
}

func (e *Endpoint) Endpoint() kitendpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		serviceResponse, err := e.Service.Get(ctx, version.Request{})
		if err != nil {
			return nil, microerror.Mask(err)
		}

		response := &Response{
			Description:    serviceResponse.Description,
			GitCommit:      serviceResponse.GitCommit,
			GoVersion:      serviceResponse.GoVersion,
			Name:           serviceResponse.Name,
			OSArch:         serviceResponse.OSArch,
			Source:         serviceResponse.Source,
			Version:        serviceResponse.Version,
			VersionBundles: serviceResponse.VersionBundles,
		}

		return response, nil
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
