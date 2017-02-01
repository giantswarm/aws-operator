// Package server provides a server implementation to connect network transport
// protocols and service business logic by defining server endpoints.
package server

import (
	"net/http"
	"sync"

	microerror "github.com/giantswarm/microkit/error"
	micrologger "github.com/giantswarm/microkit/logger"
	microserver "github.com/giantswarm/microkit/server"
	microtransaction "github.com/giantswarm/microkit/transaction"
	kithttp "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
	"golang.org/x/net/context"

	"github.com/giantswarm/aws-operator/server/endpoint"
	"github.com/giantswarm/aws-operator/server/middleware"
	"github.com/giantswarm/aws-operator/service"
)

// Config represents the configuration used to create a new server object.
type Config struct {
	// Dependencies.
	Logger               micrologger.Logger
	Router               *mux.Router
	Service              *service.Service
	TransactionResponder microtransaction.Responder

	// Settings.
	ServiceName string
}

// DefaultConfig provides a default configuration to create a new server object
// by best effort.
func DefaultConfig() Config {
	var err error

	var transactionResponder microtransaction.Responder
	{
		transactionConfig := microtransaction.DefaultResponderConfig()
		transactionResponder, err = microtransaction.NewResponder(transactionConfig)
		if err != nil {
			panic(err)
		}
	}

	config := Config{
		// Dependencies.
		Logger:               nil,
		Router:               mux.NewRouter(),
		Service:              nil,
		TransactionResponder: transactionResponder,

		// Settings.
		ServiceName: "",
	}

	return config
}

// New creates a new configured server object.
func New(config Config) (microserver.Server, error) {
	// Dependencies.
	if config.Logger == nil {
		return nil, microerror.MaskAnyf(invalidConfigError, "logger must not be empty")
	}
	if config.Router == nil {
		return nil, microerror.MaskAnyf(invalidConfigError, "router must not be empty")
	}
	if config.TransactionResponder == nil {
		return nil, microerror.MaskAnyf(invalidConfigError, "transaction responder must not be empty")
	}

	// Dependencies.
	if config.ServiceName == "" {
		return nil, microerror.MaskAnyf(invalidConfigError, "service name must not be empty")
	}

	var err error

	var middlewareCollection *middleware.Middleware
	{
		middlewareConfig := middleware.DefaultConfig()
		middlewareConfig.Logger = config.Logger
		middlewareConfig.Service = config.Service
		middlewareCollection, err = middleware.New(middlewareConfig)
		if err != nil {
			return nil, microerror.MaskAny(err)
		}
	}

	var endpointCollection *endpoint.Endpoint
	{
		endpointConfig := endpoint.DefaultConfig()
		endpointConfig.Logger = config.Logger
		endpointConfig.Middleware = middlewareCollection
		endpointConfig.Service = config.Service
		endpointCollection, err = endpoint.New(endpointConfig)
		if err != nil {
			return nil, microerror.MaskAny(err)
		}
	}

	newServer := &server{
		// Dependencies.
		logger:               config.Logger,
		router:               config.Router,
		transactionResponder: config.TransactionResponder,

		// Internals
		bootOnce: sync.Once{},
		endpoints: []microserver.Endpoint{
			endpointCollection.Version,
		},
		shutdownOnce: sync.Once{},

		// Settings.
		serviceName: config.ServiceName,
	}

	return newServer, nil
}

type server struct {
	// Dependencies.
	logger               micrologger.Logger
	router               *mux.Router
	transactionResponder microtransaction.Responder

	// Internals.
	bootOnce     sync.Once
	endpoints    []microserver.Endpoint
	shutdownOnce sync.Once

	// Settings.
	serviceName string
}

func (s *server) Boot() {
	s.bootOnce.Do(func() {
		// Here goes your custom boot logic for your server/endpoint/middleware, if
		// any.
	})
}

func (s *server) Endpoints() []microserver.Endpoint {
	return s.endpoints
}

// ErrorEncoder is a global error handler used for all endpoints. Errors
// received here are encoded by go-kit and express in which area the error was
// emitted. The underlying error defines the HTTP status code and the encoded
// error message. The response is always a JSON object containing an error field
// describing the error.
func (s *server) ErrorEncoder() kithttp.ErrorEncoder {
	return func(ctx context.Context, err error, w http.ResponseWriter) {
		switch e := err.(type) {
		case kithttp.Error:
			err = e.Err

			switch e.Domain {
			case kithttp.DomainEncode:
				w.WriteHeader(http.StatusBadRequest)
			case kithttp.DomainDecode:
				w.WriteHeader(http.StatusBadRequest)
			case kithttp.DomainDo:
				w.WriteHeader(http.StatusBadRequest)
			default:
				w.WriteHeader(http.StatusInternalServerError)
			}
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
	}
}

func (s *server) Logger() micrologger.Logger {
	return s.logger
}

func (s *server) RequestFuncs() []kithttp.RequestFunc {
	return []kithttp.RequestFunc{
		func(ctx context.Context, r *http.Request) context.Context {
			// Your custom logic to enrich the request context with request specific
			// information goes here.
			return ctx
		},
	}
}

func (s *server) Router() *mux.Router {
	return s.router
}

func (s *server) ServiceName() string {
	return s.serviceName
}

func (s *server) Shutdown() {
	s.shutdownOnce.Do(func() {
		// Here goes your custom shutdown logic for your server/endpoint/middleware,
		// if any.
	})
}

func (s *server) TransactionResponder() microtransaction.Responder {
	return s.transactionResponder
}
