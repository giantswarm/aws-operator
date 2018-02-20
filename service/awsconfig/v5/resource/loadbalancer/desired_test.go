package loadbalancer

import (
	"context"
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

	result, err := newResource.GetDesiredState(context.TODO(), customObject)
	if err != nil {
		t.Errorf("unexpected error %v", err)
	}

	if result != nil {
		t.Errorf("expected desired state '%v', got '%v'", nil, result)
	}
}
