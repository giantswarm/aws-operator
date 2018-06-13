// Package server provides a server implementation to connect network transport
// protocols and service business logic by defining server endpoints.
package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/giantswarm/micrologger"
	kitendpoint "github.com/go-kit/kit/endpoint"
	kithttp "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/viper"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/microkit/tls"
)

type Config struct {
	// ErrorEncoder is the server's error encoder. This wraps the error encoder
	// configured by the client. Clients should not implement error logging in
	// here them self. This is done by the server itself. Clients must not
	// implement error response writing them self. This is done by the server
	// itself. Duplicated response writing will lead to runtime panics.
	ErrorEncoder kithttp.ErrorEncoder
	// Logger is the logger used to print log messages.
	Logger micrologger.Logger
	// Router is a HTTP handler for the server. The returned router will have all
	// endpoints registered that are listed in the endpoint collection.
	Router *mux.Router

	// Endpoints is the server's configured list of endpoints. These are the
	// custom endpoints configured by the client.
	Endpoints []Endpoint
	// HandlerWrapper is a wrapper provided to interact with the request on its
	// roots.
	HandlerWrapper func(h http.Handler) http.Handler
	// ListenAddress is the address the server is listening on.
	ListenAddress string
	// ListenMetricsAddress is an optional address where the server will expose the
	// `/metrics` endpoint for prometheus scraping. When left blank the `/metrics`
	// endpoint will be available at the ListenAddress.
	ListenMetricsAddress string
	// LogAccess decides whether to emit logs for each requested route.
	LogAccess bool
	// RequestFuncs is the server's configured list of request functions. These
	// are the custom request functions configured by the client.
	RequestFuncs []kithttp.RequestFunc
	// ServiceName is the name of the micro-service implementing the microkit
	// server. This is used for logging and instrumentation.
	ServiceName string
	// TLSCAFile is the file path to the certificate root CA file, if any.
	TLSCAFile string
	// TLSKeyFilePath is the file path to the certificate public key file, if any.
	TLSCrtFile string
	// TLSKeyFilePath is the file path to the certificate private key file, if
	// any.
	TLSKeyFile string
	// Viper is a configuration management object.
	Viper *viper.Viper
}

// New creates a new configured server object.
func New(config Config) (Server, error) {
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "logger must not be empty")
	}
	if config.Router == nil {
		config.Router = mux.NewRouter()
	}

	if config.Endpoints == nil {
		return nil, microerror.Maskf(invalidConfigError, "endpoints must not be empty")
	}
	if config.ErrorEncoder == nil {
		config.ErrorEncoder = func(ctx context.Context, serverError error, w http.ResponseWriter) {}
	}
	if config.HandlerWrapper == nil {
		config.HandlerWrapper = func(h http.Handler) http.Handler { return h }
	}
	if config.ListenAddress == "" {
		return nil, microerror.Maskf(invalidConfigError, "listen address must not be empty")
	}
	if config.RequestFuncs == nil {
		config.RequestFuncs = []kithttp.RequestFunc{}
	}
	if config.ServiceName == "" {
		config.ServiceName = "microkit"
	}
	if config.TLSCrtFile == "" && config.TLSKeyFile != "" {
		return nil, microerror.Maskf(invalidConfigError, "TLS public key must not be empty")
	}
	if config.TLSCrtFile != "" && config.TLSKeyFile == "" {
		return nil, microerror.Maskf(invalidConfigError, "TLS private key must not be empty")
	}
	if config.Viper == nil {
		config.Viper = viper.New()
	}

	listenURL, err := url.Parse(config.ListenAddress)
	if err != nil {
		return nil, microerror.Maskf(invalidConfigError, err.Error())
	}

	var listenMetricsURL *url.URL
	if config.ListenMetricsAddress != "" {
		listenMetricsURL, err = url.Parse(config.ListenMetricsAddress)
		if err != nil {
			return nil, microerror.Maskf(invalidConfigError, err.Error())
		}

		// Check if the user supplied a https scheme for the optional metrics endpoint
		// listener. Let them know that tls configuration for this endpoint is not yet
		// implemented.
		if listenMetricsURL.Scheme == "https" {
			return nil, microerror.Maskf(invalidConfigError, "The optional metrics listener currently does not support tls configuration. Listening on https is thus disabled.")
		}
	}

	newServer := &server{
		errorEncoder: config.ErrorEncoder,
		logger:       config.Logger,
		router:       config.Router,

		bootOnce:          sync.Once{},
		config:            config,
		httpServer:        nil,
		metricsHTTPServer: nil,
		listenURL:         listenURL,
		listenMetricsUrl:  listenMetricsURL,
		shutdownOnce:      sync.Once{},

		endpoints:      config.Endpoints,
		handlerWrapper: config.HandlerWrapper,
		logAccess:      config.LogAccess,
		requestFuncs:   config.RequestFuncs,
		serviceName:    config.ServiceName,
		tlsCertFiles: tls.CertFiles{
			RootCAs: []string{config.TLSCAFile},
			Cert:    config.TLSCrtFile,
			Key:     config.TLSKeyFile,
		},
	}

	return newServer, nil
}

// server manages the transport logic and endpoint registration.
type server struct {
	// Dependencies.
	errorEncoder kithttp.ErrorEncoder
	logger       micrologger.Logger
	router       *mux.Router

	// Internals.
	bootOnce          sync.Once
	config            Config
	httpServer        *http.Server
	metricsHTTPServer *http.Server
	listenURL         *url.URL
	listenMetricsUrl  *url.URL
	shutdownOnce      sync.Once

	// Settings.
	endpoints      []Endpoint
	handlerWrapper func(h http.Handler) http.Handler
	logAccess      bool
	requestFuncs   []kithttp.RequestFunc
	serviceName    string
	tlsCertFiles   tls.CertFiles
}

func (s *server) Boot() {
	s.bootOnce.Do(func() {
		s.router.NotFoundHandler = s.newNotFoundHandler()

		// We go through all endpoints this server defines and register them to the
		// router.
		for _, e := range s.endpoints {
			func(e Endpoint) {
				// Register all endpoints to the router depending on their HTTP methods and
				// request paths. The registered http.Handler is instrumented using
				// prometheus. We track counts of execution and duration it took to complete
				// the http.Handler.
				s.router.Methods(e.Method()).Path(e.Path()).Handler(s.handlerWrapper(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					ctx, err := s.newRequestContext(w, r)
					if err != nil {
						s.newErrorEncoderWrapper()(ctx, err, w)
						return
					}

					responseWriter, err := s.newResponseWriter(w)
					if err != nil {
						s.newErrorEncoderWrapper()(ctx, err, w)
						return
					}

					// Here we define the metrics labels. These will be used to instrument
					// the current request. This defered callback is initialized with the
					// timestamp of the beginning of the execution and will be executed at
					// the very end of the request. When it is executed we know all
					// necessary information to instrument the complete request, including
					// its response status code.
					defer func(t time.Time) {
						endpointCode := strconv.Itoa(responseWriter.StatusCode())
						endpointMethod := strings.ToLower(e.Method())
						endpointName := strings.Replace(e.Name(), "/", "_", -1)

						if s.logAccess {
							s.logger.Log("code", endpointCode, "endpoint", e.Name(), "level", "debug", "message", "tracking access log", "method", endpointMethod, "path", r.URL.Path)
						}

						endpointTotal.WithLabelValues(endpointCode, endpointMethod, endpointName).Inc()
						endpointTime.WithLabelValues(endpointCode, endpointMethod, endpointName).Set(float64(time.Since(t) / time.Millisecond))
					}(time.Now())

					// Combine all options this server defines. Since the interface of the
					// go-kit server changed to not accept a context anymore we have to
					// work around the context injection by injecting our context via the
					// very first request function.
					//
					// NOTE this is rather an ugly hack and should be revisited. It would
					// probably make sense to start decoupling from the go-kit code since
					// there haven't been any benefits from its implementation, but only
					// from its design ideas. Also note that some of the design ideas
					// dictated by go-kit do not align with our own ideas and often stood
					// in our way of making things work how they should be.
					options := []kithttp.ServerOption{
						kithttp.ServerBefore(func(context.Context, *http.Request) context.Context {
							return ctx
						}),
						kithttp.ServerBefore(s.requestFuncs...),
						kithttp.ServerErrorEncoder(s.newErrorEncoderWrapper()),
					}

					// Now we execute the actual go-kit endpoint handler.
					kithttp.NewServer(
						s.newEndpointWrapper(e),
						e.Decoder(),
						e.Encoder(),
						options...,
					).ServeHTTP(responseWriter, r)
				})))
			}(e)
		}

		// If the user provided a specific url for the metrics endpoint:
		if s.listenMetricsUrl != nil {
			// Register prometheus metrics endpoint to a different server as the rest
			// of the endpoints.
			go func() {
				s.logger.Log("level", "debug", "message", fmt.Sprintf("running metrics server at %s", s.listenMetricsUrl.String()))

				metricsRouter := mux.NewRouter()
				metricsRouter.Path("/metrics").Handler(promhttp.Handler())

				s.metricsHTTPServer = &http.Server{
					Addr:    s.listenMetricsUrl.Host,
					Handler: metricsRouter,
				}

				err := s.metricsHTTPServer.ListenAndServe()
				if IsServerClosed(err) {
					// We get a closed error in case the server is shutting down. We expect
					// this at times so we just fall through here.
				} else if err != nil {
					panic(err)
				}
			}()
		} else {
			// Register prometheus metrics endpoint to the same server as the rest of
			// the endpoints.
			s.router.Path("/metrics").Handler(promhttp.Handler())
		}

		// Register the router which has all of the configured custom endpoints
		// registered.
		s.httpServer = &http.Server{
			Addr:    s.listenURL.Host,
			Handler: s.router,
		}

		go func() {
			s.logger.Log("level", "debug", "message", fmt.Sprintf("running server at %s", s.listenURL.String()))

			if s.listenURL.Scheme == "https" {
				tlsConfig, err := tls.LoadTLSConfig(s.tlsCertFiles)
				if err != nil {
					panic(err)
				}
				s.httpServer.TLSConfig = tlsConfig
			}

			err := s.httpServer.ListenAndServe()
			if IsServerClosed(err) {
				// We get a closed error in case the server is shutting down. We expect
				// this at times so we just fall through here.
			} else if err != nil {
				panic(err)
			}
		}()
	})
}

func (s *server) Config() Config {
	return s.config
}

func (s *server) Shutdown() {
	s.shutdownOnce.Do(func() {
		// Stop the HTTP server gracefully and wait some time for open connections
		// to be closed. Then force it to be stopped.
		go func() {
			err := s.httpServer.Shutdown(context.Background())
			if err != nil {
				s.logger.Log("level", "error", "message", "shutting down server failed", "stack", fmt.Sprintf("%#v", err))
			}
		}()
		<-time.After(3 * time.Second)
		err := s.httpServer.Close()
		if err != nil {
			s.logger.Log("level", "error", "message", "closing server failed", "stack", fmt.Sprintf("%#v", err))
		}
	})
}

// newEndpointWrapper creates a new wrapped endpoint function essentially
// combining the actual endpoint implementation with the defined middlewares.
func (s *server) newEndpointWrapper(e Endpoint) kitendpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		// Prepare the actual endpoint depending on the provided middlewares of the
		// endpoint implementation. There might be cases in which there are none or
		// only one middleware. The go-kit interface is not that nice so we need to
		// make it fit here.
		endpoint := e.Endpoint()
		middlewares := e.Middlewares()
		if len(middlewares) == 1 {
			endpoint = kitendpoint.Chain(middlewares[0])(endpoint)
		}
		if len(middlewares) > 1 {
			endpoint = kitendpoint.Chain(middlewares[0], middlewares[1:]...)(endpoint)
		}
		response, err := endpoint(ctx, request)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		return response, nil
	}
}

func (s *server) newErrorEncoderWrapper() kithttp.ErrorEncoder {
	return func(ctx context.Context, serverError error, w http.ResponseWriter) {
		var err error

		// At first we have to set the content type of the actual error response. If
		// we would set it at the end we would set a trailing header that would not
		// be recognized by most of the clients out there. This is because in the
		// next call to the errorEncoder below the client's implementation of the
		// errorEncoder probably writes the status code header, which marks the
		// beginning of trailing headers in HTTP.
		w.Header().Set("Content-Type", "application/json; charset=utf-8")

		// Create the microkit specific response error, which acts as error wrapper
		// within the client's error encoder. It is used to propagate response codes
		// and messages, so we can use them below.
		var responseError ResponseError
		{
			responseConfig := DefaultResponseErrorConfig()
			responseConfig.Underlying = serverError
			responseError, err = NewResponseError(responseConfig)
			if err != nil {
				panic(err)
			}
		}

		rw, err := s.newResponseWriter(w)
		if err != nil {
			panic(err)
		}

		// Run the custom error encoder. This is used to let the implementing
		// microservice do something with errors occured during runtime. Things like
		// writing specific HTTP status codes to the given response writer or
		// writing data to the response body can be done.
		s.errorEncoder(ctx, responseError, rw)

		// Log the error and its stack. This is really useful for debugging.
		s.logger.Log("level", "error", "message", "stop endpoint processing due to error", "stack", fmt.Sprintf("%#v", serverError))

		// Emit metrics about the occured errors. That way we can feed our
		// instrumentation stack to have nice dashboards to get a picture about the
		// general system health.
		errorTotal.WithLabelValues().Inc()

		// Write the actual response body in case no response was already written
		// inside the error encoder.
		if !rw.HasWritten() {
			json.NewEncoder(rw).Encode(map[string]interface{}{
				"code":  responseError.Code(),
				"error": responseError.Message(),
				"from":  s.serviceName,
			})
		}
	}
}

// newNotFoundHandler returns an HTTP handler that represents our custom not
// found handler. Here we take care about logging, metrics and a proper
// response.
func (s *server) newNotFoundHandler() http.Handler {
	return http.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		errMessage := fmt.Sprintf("endpoint not found for %s %s", r.Method, r.URL.Path)

		// Log the error and its message. This is really useful for debugging.
		s.logger.Log("level", "error", "message", errMessage)

		// This defered callback will be executed at the very end of the request.
		defer func(t time.Time) {
			endpointCode := strconv.Itoa(http.StatusNotFound)
			endpointMethod := strings.ToLower(r.Method)
			endpointName := "notfound"

			endpointTotal.WithLabelValues(endpointCode, endpointMethod, endpointName).Inc()
			endpointTime.WithLabelValues(endpointCode, endpointMethod, endpointName).Set(float64(time.Since(t) / time.Millisecond))

			errorTotal.WithLabelValues().Inc()
		}(time.Now())

		// Write the actual response body.
		w.WriteHeader(http.StatusNotFound)
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"code":  CodeResourceNotFound,
			"error": errMessage,
			"from":  s.serviceName,
		})
	}))
}

// newRequestContext creates a new request context and enriches it with request
// relevant information. E.g. here we put the HTTP X-Idempotency-Key header into
// the request context, if any. We also check if there is a transaction response
// already tracked for the given transaction ID. This information is then stored
// within the given request context as well. Note that we initialize the
// information about the tracked state of the transaction response with false,
// to always have a valid state available within the request context.
func (s *server) newRequestContext(w http.ResponseWriter, r *http.Request) (context.Context, error) {
	ctx := context.Background()

	return ctx, nil
}

// newResponseWriter creates a new wrapped HTTP response writer. E.g. here we
// create a new wrapper for the http.ResponseWriter of the current request. We
// inject it into the called http.Handler so it can track the status code we are
// interested in. It will help us gathering the response status code after it
// was written by the underlying http.ResponseWriter.
func (s *server) newResponseWriter(w http.ResponseWriter) (ResponseWriter, error) {
	responseConfig := DefaultResponseWriterConfig()
	responseConfig.ResponseWriter = w
	responseWriter, err := NewResponseWriter(responseConfig)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return responseWriter, nil
}
