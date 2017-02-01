package version

import (
	"encoding/json"
	"net/http"

	microerror "github.com/giantswarm/microkit/error"
	micrologger "github.com/giantswarm/microkit/logger"
	kitendpoint "github.com/go-kit/kit/endpoint"
	kithttp "github.com/go-kit/kit/transport/http"
	"golang.org/x/net/context"

	"github.com/giantswarm/aws-operator/server/middleware"
	"github.com/giantswarm/aws-operator/service"
	"github.com/giantswarm/aws-operator/service/version"
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
	Logger     micrologger.Logger
	Middleware *middleware.Middleware
	Service    *service.Service
}

// DefaultConfig provides a default configuration to create a new version
// endpoint by best effort.
func DefaultConfig() Config {
	return Config{
		// Dependencies.
		Logger:     nil,
		Middleware: nil,
		Service:    nil,
	}
}

// New creates a new configured version endpoint.
func New(config Config) (*Endpoint, error) {
	// Dependencies.
	if config.Logger == nil {
		return nil, microerror.MaskAnyf(invalidConfigError, "logger must not be empty")
	}
	if config.Middleware == nil {
		return nil, microerror.MaskAnyf(invalidConfigError, "middleware must not be empty")
	}
	if config.Service == nil {
		return nil, microerror.MaskAnyf(invalidConfigError, "service must not be empty")
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
		serviceResponse, err := e.Service.Version.Get(ctx, version.DefaultRequest())
		if err != nil {
			return nil, microerror.MaskAny(err)
		}

		response := DefaultResponse()
		response.Description = serviceResponse.Description
		response.GitCommit = serviceResponse.GitCommit
		response.GoVersion = serviceResponse.GoVersion
		response.Name = serviceResponse.Name
		response.OSArch = serviceResponse.OSArch
		response.Source = serviceResponse.Source

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
