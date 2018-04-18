package endpoints

import (
	"context"
	"testing"

	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/micrologger/microloggertest"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func Test_Resource_Endpoints_GetDesiredState(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		description       string
		obj               interface{}
		expectedNamespace string
		expectedName      string
		expectedIPAddress string
		expectedPort      int
	}{
		{
			description: "basic match",
			obj: &v1alpha1.AWSConfig{
				Spec: v1alpha1.AWSConfigSpec{
					Cluster: v1alpha1.Cluster{
						ID: "al9qy",
					},
				},
			},
			expectedNamespace: "al9qy",
			expectedName:      "master",
			expectedIPAddress: "10.1.1.1",
			expectedPort:      443,
		},
	}

	var err error
	var newResource *Resource

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			c := Config{
				Clients: Clients{
					EC2: &EC2ClientMock{
						privateIPAddress: tc.expectedIPAddress,
					},
				},
				K8sClient: fake.NewSimpleClientset(),
				Logger:    microloggertest.New(),
			}
			newResource, err = New(c)

			if err != nil {
				t.Fatal("expected", nil, "got", err)
			}

			result, err := newResource.GetDesiredState(context.TODO(), tc.obj)
			if err != nil {
				t.Fatalf("expected '%v' got '%#v'", nil, err)
			}
			desiredEndpoints, ok := result.(*apiv1.Endpoints)
			if !ok {
				t.Fatalf("case expected '%T', got '%T'", desiredEndpoints, result)
			}

			if tc.expectedNamespace != desiredEndpoints.ObjectMeta.Namespace {
				t.Fatalf("expected namespace %q got %q", tc.expectedNamespace, desiredEndpoints.ObjectMeta.Namespace)
			}

			if tc.expectedName != desiredEndpoints.ObjectMeta.Name {
				t.Fatalf("expected name %q got %q", tc.expectedName, desiredEndpoints.ObjectMeta.Name)
			}

			if int32(tc.expectedPort) != desiredEndpoints.Subsets[0].Ports[0].Port {
				t.Fatalf("expected port %q got %q", int32(tc.expectedPort), desiredEndpoints.Subsets[0].Ports[0].Port)
			}

			if tc.expectedIPAddress != desiredEndpoints.Subsets[0].Addresses[0].IP {
				t.Fatalf("expected ip address %s got %s", tc.expectedIPAddress, desiredEndpoints.Subsets[0].Addresses[0].IP)
			}
		})
	}
}
