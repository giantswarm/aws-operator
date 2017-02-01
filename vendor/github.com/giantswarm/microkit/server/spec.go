package server

import (
	"bytes"
	"net/http"

	kitendpoint "github.com/go-kit/kit/endpoint"
	kithttp "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"

	micrologger "github.com/giantswarm/microkit/logger"
	microtransaction "github.com/giantswarm/microkit/transaction"
)

// Endpoint represents the management of transport logic. An endpoint defines
// what it needs to work properly. Internally it holds a reference to the
// service object which implements business logic and executes any workload.
// That means that network transport and business logic are strictly separated
// and work hand in hand via well defined package APIs.
type Endpoint interface {
	// Decoder returns the kithttp.DecodeRequestFunc used to decode a request
	// before the actual endpoint is executed.
	Decoder() kithttp.DecodeRequestFunc
	// Decoder returns the kithttp.EncodeResponseFunc used to encode a response
	// after the actual endpoint was executed.
	Encoder() kithttp.EncodeResponseFunc
	// Endpoint returns the kitendpoint.Endpoint which receives a decoded response
	// and forwards any workload to the internal service object reference.
	Endpoint() kitendpoint.Endpoint
	// Method returns the HTTP verb used to register the endpoint.
	Method() string
	// Middlewares returns the middlewares the endpoint configures to be run
	// before the actual endpoint is being invoked.
	Middlewares() []kitendpoint.Middleware
	// Name returns the name of the endpoint which can be used to label metrics or
	// annotate logs.
	Name() string
	// Path returns the HTTP request URL path used to register the endpoint.
	Path() string
}

// Server manages the HTTP transport logic.
type Server interface {
	// Boot registers the configured endpoints and starts the server under the
	// configured address.
	Boot()
	// Endpoints returns the server's configured list of endpoints. These are the
	// custom endpoints configured by the client.
	Endpoints() []Endpoint
	// ErrorEncoder returns the server's error encoder. This wraps the error
	// encoder configured by the client. Clients should not implement error
	// logging in here them self. This is done by the server itself. Clients must
	// not implement error response writing them self. This is done by the server
	// itself. Duplicated response writing will lead to runtime panics.
	ErrorEncoder() kithttp.ErrorEncoder
	Logger() micrologger.Logger
	// RequestFuncs returns the server's configured list of request functions.
	// These are the custom request functions configured by the client.
	RequestFuncs() []kithttp.RequestFunc
	// Router returns a HTTP handler for the server. The returned router will have
	// all endpoints registered that are listed in the endpoint collection.
	Router() *mux.Router
	// ServiceName returns the name of the micro-service implementing the microkit
	// server. This is used for logging and instrumentation.
	ServiceName() string
	// Shutdown stops the running server gracefully.
	Shutdown()
	TransactionResponder() microtransaction.Responder
}

// ResponseError is a wrapper for errors passed to the client's error encoder to
// propagate specific response information in error cases.
type ResponseError interface {
	// Code returns the code being tracked using SetCode. If this code is not set
	// using SetCode it defaults to CodeUnknownError.
	Code() string
	// Error returns the message of the underlying error.
	Error() string
	// Message returns the message being tracked using SetMessage. If this message
	// is not set using SetMessage it defaults to the error message of the
	// underlying error.
	Message() string
	// IsEndpoint checks whether the underlying error originates from an endpoints
	// business logic. This includes decoder and encoder errors. In case
	// IsEndpoint returns false, something unexpected happened and the current
	// error should probably be handled as internal server error.
	IsEndpoint() bool
	// SetCode tracks the given response code for the current response error. The
	// given response code will be used for logging, instrumentation and response
	// creation.
	SetCode(code string)
	// SetMessage tracks the given response message for the current response
	// error. The given response message will be used for response creation.
	SetMessage(message string)
	// Underlying returns the actual underlying error, which is expected to be of
	// type kithttp.Error.
	Underlying() error
}

// ResponseWriter is a wrapper for http.ResponseWriter to track the written
// status code.
type ResponseWriter interface {
	// BodyBuffer returns the buffer which is used to track the bytes being
	// written to the response.
	BodyBuffer() *bytes.Buffer
	// Header is only a wrapper around http.ResponseWriter.Header.
	Header() http.Header
	// StatusCode returns either the default status code of the one that was
	// actually written using WriteHeader.
	StatusCode() int
	// Write is only a wrapper around http.ResponseWriter.Write.
	Write(b []byte) (int, error)
	// WriteHeader is a wrapper around http.ResponseWriter.Write. In addition to
	// that it is used to track the written status code.
	WriteHeader(c int)
}
