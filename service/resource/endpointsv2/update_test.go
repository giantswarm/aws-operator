package endpointsv2

import (
	"context"
	"testing"

	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/micrologger/microloggertest"
	apismetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	apiv1 "k8s.io/api/core/v1"
)

func Test_Resource_Endpoints_newUpdateChange(t *testing.T) {
	testCases := []struct {
		description       string
		obj               interface{}
		cur               interface{}
		des               interface{}
		expectedEndpoints *apiv1.Endpoints
	}{
		{
			description: "current and desired states are different, return desired state",
			obj: &v1alpha1.AWSConfig{
				Spec: v1alpha1.AWSConfigSpec{
					Cluster: v1alpha1.Cluster{
						ID: "foobar",
					},
				},
			},
			cur: &apiv1.Endpoints{
				TypeMeta: apismetav1.TypeMeta{
					Kind:       "Endpoints",
					APIVersion: "v1",
				},
				ObjectMeta: apismetav1.ObjectMeta{
					Name: "master",
				},
				Subsets: []apiv1.EndpointSubset{
					{
						Addresses: []apiv1.EndpointAddress{
							{
								IP: "10.0.0.1",
							},
						},
					},
				},
			},
			des: &apiv1.Endpoints{
				TypeMeta: apismetav1.TypeMeta{
					Kind:       "Endpoints",
					APIVersion: "v1",
				},
				ObjectMeta: apismetav1.ObjectMeta{
					Name: "master",
				},
				Subsets: []apiv1.EndpointSubset{
					{
						Addresses: []apiv1.EndpointAddress{
							{
								IP: "127.0.0.1",
							},
						},
					},
				},
			},
			expectedEndpoints: &apiv1.Endpoints{
				TypeMeta: apismetav1.TypeMeta{
					Kind:       "Endpoints",
					APIVersion: "v1",
				},
				ObjectMeta: apismetav1.ObjectMeta{
					Name: "master",
				},
				Subsets: []apiv1.EndpointSubset{
					{
						Addresses: []apiv1.EndpointAddress{
							{
								IP: "127.0.0.1",
							},
						},
					},
				},
			},
		},
		{
			description: "current state matches desired state, no change",
			obj: &v1alpha1.AWSConfig{
				Spec: v1alpha1.AWSConfigSpec{
					Cluster: v1alpha1.Cluster{
						ID: "foobar",
					},
				},
			},
			cur: &apiv1.Endpoints{
				TypeMeta: apismetav1.TypeMeta{
					Kind:       "Endpoints",
					APIVersion: "v1",
				},
				ObjectMeta: apismetav1.ObjectMeta{
					Name: "master",
				},
				Subsets: []apiv1.EndpointSubset{
					{
						Addresses: []apiv1.EndpointAddress{
							{
								IP: "10.0.0.1",
							},
						},
					},
				},
			},
			des: &apiv1.Endpoints{
				TypeMeta: apismetav1.TypeMeta{
					Kind:       "Endpoints",
					APIVersion: "v1",
				},
				ObjectMeta: apismetav1.ObjectMeta{
					Name: "master",
				},
				Subsets: []apiv1.EndpointSubset{
					{
						Addresses: []apiv1.EndpointAddress{
							{
								IP: "10.0.0.1",
							},
						},
					},
				},
			},
			expectedEndpoints: nil,
		},
		{
			description: "current state is empty, no change",
			obj: &v1alpha1.AWSConfig{
				Spec: v1alpha1.AWSConfigSpec{
					Cluster: v1alpha1.Cluster{
						ID: "foobar",
					},
				},
			},
			cur: nil,
			des: &apiv1.Endpoints{
				TypeMeta: apismetav1.TypeMeta{
					Kind:       "Endpoints",
					APIVersion: "v1",
				},
				ObjectMeta: apismetav1.ObjectMeta{
					Name: "master",
				},
				Subsets: []apiv1.EndpointSubset{
					{
						Addresses: []apiv1.EndpointAddress{
							{
								IP: "10.0.0.1",
							},
						},
					},
				},
			},
			expectedEndpoints: nil,
		},
	}

	var err error
	var newResource *Resource
	{
		resourceConfig := DefaultConfig()
		resourceConfig.K8sClient = fake.NewSimpleClientset()
		resourceConfig.Logger = microloggertest.New()
		newResource, err = New(resourceConfig)
		if err != nil {
			t.Fatal("expected", nil, "got", err)
		}
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			result, err := newResource.newUpdateChange(context.TODO(), tc.obj, tc.cur, tc.des)
			if err != nil {
				t.Errorf("expected '%v' got '%#v'", nil, err)
			}
			if tc.expectedEndpoints == nil {
				if tc.expectedEndpoints != result {
					t.Errorf("expected '%v' got '%v'", tc.expectedEndpoints, result)
				}
			} else {
				endpointsToUpdate, ok := result.(*apiv1.Endpoints)
				if !ok {
					t.Errorf("case expected '%T', got '%T'", endpointsToUpdate, result)
				}
				if tc.expectedEndpoints.Name != endpointsToUpdate.Name {
					t.Errorf("expected %s, got %s", tc.expectedEndpoints.Name, endpointsToUpdate.Name)
				}
			}
		})
	}
}
