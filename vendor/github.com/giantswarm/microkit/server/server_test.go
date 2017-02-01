package server

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	kitendpoint "github.com/go-kit/kit/endpoint"
	kithttp "github.com/go-kit/kit/transport/http"
	"golang.org/x/net/context"
)

func Test_Transaction_NoIDGiven(t *testing.T) {
	e := testNewEndpoint(t)

	config := DefaultConfig()
	config.Endpoints = []Endpoint{e}
	newServer, err := New(config)
	if err != nil {
		t.Fatal("expected", nil, "got", err)
	}

	newServer.Boot()
	defer newServer.Shutdown()

	// Here we make a request against our test endpoint. The endpoint is executed
	// the first time. So the execution counts should be one.
	{
		r, err := http.NewRequest("GET", "/test-path", nil)
		if err != nil {
			t.Fatal("expected", nil, "got", err)
		}
		w := httptest.NewRecorder()

		newServer.Router().ServeHTTP(w, r)

		if w.Code != http.StatusOK {
			t.Fatal("expected", http.StatusOK, "got", w.Code)
		}

		decoderExecuted := e.(*testEndpoint).decoderExecuted
		if decoderExecuted != 1 {
			t.Fatal("expected", 1, "got", decoderExecuted)
		}
		endpointExecuted := e.(*testEndpoint).endpointExecuted
		if endpointExecuted != 1 {
			t.Fatal("expected", 1, "got", endpointExecuted)
		}
		encoderExecuted := e.(*testEndpoint).encoderExecuted
		if encoderExecuted != 1 {
			t.Fatal("expected", 1, "got", encoderExecuted)
		}

		if w.Body.String() != "test-response-1" {
			t.Fatal("expected", "test-response-1", "got", w.Body.String())
		}
	}

	// Here we make another request against our test endpoint. The endpoint is
	// executed the second time. So the execution counts should be two.
	{
		r, err := http.NewRequest("GET", "/test-path", nil)
		if err != nil {
			t.Fatal("expected", nil, "got", err)
		}
		w := httptest.NewRecorder()

		newServer.Router().ServeHTTP(w, r)

		if w.Code != http.StatusOK {
			t.Fatal("expected", http.StatusOK, "got", w.Code)
		}

		decoderExecuted := e.(*testEndpoint).decoderExecuted
		if decoderExecuted != 2 {
			t.Fatal("expected", 2, "got", decoderExecuted)
		}
		endpointExecuted := e.(*testEndpoint).endpointExecuted
		if endpointExecuted != 2 {
			t.Fatal("expected", 2, "got", endpointExecuted)
		}
		encoderExecuted := e.(*testEndpoint).encoderExecuted
		if encoderExecuted != 2 {
			t.Fatal("expected", 2, "got", encoderExecuted)
		}

		if w.Body.String() != "test-response-2" {
			t.Fatal("expected", "test-response-2", "got", w.Body.String())
		}
	}
}

func Test_Transaction_IDGiven(t *testing.T) {
	e := testNewEndpoint(t)

	config := DefaultConfig()
	config.Endpoints = []Endpoint{e}
	newServer, err := New(config)
	if err != nil {
		t.Fatal("expected", nil, "got", err)
	}

	newServer.Boot()
	defer newServer.Shutdown()

	// Here we make a request against our test endpoint. The endpoint is executed
	// the first time. So the execution counts should be one.
	{
		r, err := http.NewRequest("GET", "/test-path", nil)
		if err != nil {
			t.Fatal("expected", nil, "got", err)
		}
		r.Header.Add(TransactionIDHeader, "my-very-valid-test-transaction-id")
		w := httptest.NewRecorder()

		newServer.Router().ServeHTTP(w, r)

		if w.Code != http.StatusOK {
			t.Fatal("expected", http.StatusOK, "got", w.Code)
		}

		decoderExecuted := e.(*testEndpoint).decoderExecuted
		if decoderExecuted != 1 {
			t.Fatal("expected", 1, "got", decoderExecuted)
		}
		endpointExecuted := e.(*testEndpoint).endpointExecuted
		if endpointExecuted != 1 {
			t.Fatal("expected", 1, "got", endpointExecuted)
		}
		encoderExecuted := e.(*testEndpoint).encoderExecuted
		if encoderExecuted != 1 {
			t.Fatal("expected", 1, "got", encoderExecuted)
		}

		if w.Body.String() != "test-response-1" {
			t.Fatal("expected", "test-response-1", "got", w.Body.String())
		}
	}

	// Here we make another request against our test endpoint. In this request and
	// the previous one we provided the same transaction ID. The endpoint is now
	// being executed the second time. So because we have our transaction response
	// tracked, the execution counts should still be one.
	{
		r, err := http.NewRequest("GET", "/test-path", nil)
		if err != nil {
			t.Fatal("expected", nil, "got", err)
		}
		r.Header.Add(TransactionIDHeader, "my-very-valid-test-transaction-id")
		w := httptest.NewRecorder()

		newServer.Router().ServeHTTP(w, r)

		if w.Code != http.StatusOK {
			t.Fatal("expected", http.StatusOK, "got", w.Code)
		}

		decoderExecuted := e.(*testEndpoint).decoderExecuted
		if decoderExecuted != 1 {
			t.Fatal("expected", 1, "got", decoderExecuted)
		}
		endpointExecuted := e.(*testEndpoint).endpointExecuted
		if endpointExecuted != 1 {
			t.Fatal("expected", 1, "got", endpointExecuted)
		}
		encoderExecuted := e.(*testEndpoint).encoderExecuted
		if encoderExecuted != 1 {
			t.Fatal("expected", 1, "got", encoderExecuted)
		}

		if w.Body.String() != "test-response-1" {
			t.Fatal("expected", "test-response-1", "got", w.Body.String())
		}
	}
}

func Test_Transaction_InvalidIDGiven(t *testing.T) {
	e := testNewEndpoint(t)

	config := DefaultConfig()
	config.Endpoints = []Endpoint{e}
	config.ErrorEncoder = func(ctx context.Context, serverError error, w http.ResponseWriter) {
		w.WriteHeader(http.StatusInternalServerError)
	}
	newServer, err := New(config)
	if err != nil {
		t.Fatal("expected", nil, "got", err)
	}

	newServer.Boot()
	defer newServer.Shutdown()

	// Here we make a request against our test endpoint. The endpoint is provided
	// with an invalid transaction ID. The server's error encoder returns status
	// code 500 on all errors.
	{
		r, err := http.NewRequest("GET", "/test-path", nil)
		if err != nil {
			t.Fatal("expected", nil, "got", err)
		}
		r.Header.Add(TransactionIDHeader, "--my-invalid-transaction-id--")
		w := httptest.NewRecorder()

		newServer.Router().ServeHTTP(w, r)

		if w.Code != http.StatusInternalServerError {
			t.Fatal("expected", http.StatusInternalServerError, "got", w.Code)
		}

		if !strings.Contains(w.Body.String(), "invalid transaction ID: does not match") {
			t.Fatal("expected", "invalid transaction ID: does not match", "got", w.Body.String())
		}
	}
}

func testNewEndpoint(t *testing.T) Endpoint {
	newEndpoint := &testEndpoint{
		decoderExecuted:        0,
		decoderRequest:         "",
		endpointExecuted:       0,
		encoderExecuted:        0,
		endpointResponseFormat: "test-response-%d",
		method:                 "GET",
		name:                   "test-endpoint",
		path:                   "/test-path",
	}

	return newEndpoint
}

type testEndpoint struct {
	decoderExecuted        int
	decoderRequest         string
	endpointExecuted       int
	encoderExecuted        int
	endpointResponseFormat string
	method                 string
	name                   string
	path                   string
}

func (e *testEndpoint) Decoder() kithttp.DecodeRequestFunc {
	return func(ctx context.Context, r *http.Request) (interface{}, error) {
		e.decoderExecuted++
		return e.decoderRequest, nil
	}
}

func (e *testEndpoint) Encoder() kithttp.EncodeResponseFunc {
	return func(ctx context.Context, w http.ResponseWriter, response interface{}) error {
		e.encoderExecuted++
		_, err := w.Write([]byte(response.(string)))
		return err
	}
}

func (e *testEndpoint) Endpoint() kitendpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		e.endpointExecuted++
		return fmt.Sprintf(e.endpointResponseFormat, e.endpointExecuted), nil
	}
}

func (e *testEndpoint) Method() string {
	return e.method
}

func (e *testEndpoint) Middlewares() []kitendpoint.Middleware {
	return []kitendpoint.Middleware{}
}

func (e *testEndpoint) Name() string {
	return e.name
}

func (e *testEndpoint) Path() string {
	return e.path
}
