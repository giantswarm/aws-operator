package ebsvolume

import (
	"context"
	"testing"

	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/aws-operator/service/awsconfig/v10/ebs"
	"github.com/giantswarm/micrologger/microloggertest"
)

func Test_DesiredState(t *testing.T) {
	t.Parallel()
	customObject := &v1alpha1.AWSConfig{
		Spec: v1alpha1.AWSConfigSpec{
			Cluster: v1alpha1.Cluster{
				ID: "test-cluster",
			},
		},
	}

	ebsConfig := ebs.Config{
		Client: &ebs.EC2ClientMock{},
		Logger: microloggertest.New(),
	}
	ebsService, err := ebs.New(ebsConfig)
	if err != nil {
		t.Error("expected", nil, "got", err)
	}

	c := Config{
		Logger:  microloggertest.New(),
		Service: ebsService,
	}
	newResource, err := New(c)
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
