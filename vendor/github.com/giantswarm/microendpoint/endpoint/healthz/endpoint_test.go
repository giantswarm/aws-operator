package healthz

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/giantswarm/micrologger/microloggertest"
	kitendpoint "github.com/go-kit/kit/endpoint"
	kithttp "github.com/go-kit/kit/transport/http"

	"github.com/giantswarm/microendpoint/service/healthz"
)

func Test_Endpoint_ServiceFailed_True(t *testing.T) {
	testCases := []struct {
		Failed             bool
		ExpectedStatusCode int
	}{
		{
			Failed:             false,
			ExpectedStatusCode: http.StatusOK,
		},
		{
			Failed:             true,
			ExpectedStatusCode: http.StatusInternalServerError,
		},
	}

	for i, tc := range testCases {
		var encoder kithttp.EncodeResponseFunc
		var endpoint kitendpoint.Endpoint
		{
			endpointConfig := DefaultConfig()
			endpointConfig.Logger = microloggertest.New()
			endpointConfig.Services = []healthz.Service{
				&testService{Failed: tc.Failed},
			}
			newEndpoint, err := New(endpointConfig)
			if err != nil {
				t.Fatalf("test %d expected %v got %s", i+1, nil, err)
			}
			encoder = newEndpoint.Encoder()
			endpoint = newEndpoint.Endpoint()
		}

		res, err := endpoint(context.TODO(), nil)
		if err != nil {
			t.Fatalf("test %d expected %v got %s", i+1, nil, err)
		}

		w := httptest.NewRecorder()

		err = encoder(context.TODO(), w, res)
		if err != nil {
			t.Fatalf("test %d expected %v got %s", i+1, nil, err)
		}

		statusCode := w.Result().StatusCode
		if statusCode != tc.ExpectedStatusCode {
			t.Fatalf("test %d expected %d got %d", i+1, tc.ExpectedStatusCode, statusCode)
		}
	}
}

type testService struct {
	Failed bool
}

func (s *testService) GetHealthz(ctx context.Context) (healthz.Response, error) {
	response := healthz.Response{
		Failed: s.Failed,
	}

	return response, nil
}
