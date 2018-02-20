package loadbalancer

import (
	"context"
	"reflect"
	"testing"

	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/micrologger/microloggertest"
)

func Test_DesiredState(t *testing.T) {
	customObject := &v1alpha1.AWSConfig{
		Spec: v1alpha1.AWSConfigSpec{
			Cluster: v1alpha1.Cluster{
				ID: "test-cluster",
			},
		},
	}

	testCases := []struct {
		description   string
		expectedState *LoadBalancerState
	}{
		{
			description: "basic match returns empty state",
			expectedState: &LoadBalancerState{
				LoadBalancerNames: []string{},
			},
		},
	}
	var err error
	var newResource *Resource

	c := Config{
		Clients: Clients{
			ELB: &ELBClientMock{},
		},
		Logger: microloggertest.New(),
	}
	newResource, err = New(c)
	if err != nil {
		t.Error("expected", nil, "got", err)
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {

			result, err := newResource.GetDesiredState(context.TODO(), customObject)
			if err != nil {
				t.Errorf("unexpected error %v", err)
			}
			desiredState, ok := result.(*LoadBalancerState)
			if !ok {
				t.Errorf("expected '%T', got '%T'", desiredState, result)
			}

			if !reflect.DeepEqual(desiredState, tc.expectedState) {
				t.Errorf("expected desired state '%#v', got '%#v'", tc.expectedState, desiredState)
			}
		})
	}
}
