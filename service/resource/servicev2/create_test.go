package servicev2

import (
	"context"
	"testing"

	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/micrologger/microloggertest"
	apismetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	apiv1 "k8s.io/client-go/pkg/api/v1"
)

func Test_Resource_Service_newCreateChange(t *testing.T) {
	testCases := []struct {
		description     string
		obj             interface{}
		cur             interface{}
		des             interface{}
		expectedService *apiv1.Service
	}{
		{
			description: "current state matches desired state, desired state is empty",
			obj: &v1alpha1.AWSConfig{
				Spec: v1alpha1.AWSConfigSpec{
					Cluster: v1alpha1.Cluster{
						ID: "foobar",
					},
				},
			},
			cur: &apiv1.Service{
				TypeMeta: apismetav1.TypeMeta{
					Kind:       "Service",
					APIVersion: "v1",
				},
				ObjectMeta: apismetav1.ObjectMeta{
					Name: "al9qy",
				},
			},
			des: &apiv1.Service{
				TypeMeta: apismetav1.TypeMeta{
					Kind:       "Service",
					APIVersion: "v1",
				},
				ObjectMeta: apismetav1.ObjectMeta{
					Name: "al9qy",
				},
			},
			expectedService: nil,
		},

		{
			description: "current state is empty, return desired state",
			obj: &v1alpha1.AWSConfig{
				Spec: v1alpha1.AWSConfigSpec{
					Cluster: v1alpha1.Cluster{
						ID: "foobar",
					},
				},
			},
			cur: nil,
			des: &apiv1.Service{
				TypeMeta: apismetav1.TypeMeta{
					Kind:       "Service",
					APIVersion: "v1",
				},
				ObjectMeta: apismetav1.ObjectMeta{
					Name: "al9qy",
				},
			},
			expectedService: &apiv1.Service{
				TypeMeta: apismetav1.TypeMeta{
					Kind:       "Service",
					APIVersion: "v1",
				},
				ObjectMeta: apismetav1.ObjectMeta{
					Name: "al9qy",
				},
			},
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
			result, err := newResource.newCreateChange(context.TODO(), tc.obj, tc.cur, tc.des)
			if err != nil {
				t.Errorf("expected '%v' got '%#v'", nil, err)
			}
			if tc.expectedService == nil {
				if tc.expectedService != result {
					t.Errorf("expected '%v' got '%v'", tc.expectedService, result)
				}
			} else {
				serviceToCreate, ok := result.(*apiv1.Service)
				if !ok {
					t.Errorf("case expected '%T', got '%T'", serviceToCreate, result)
				}
				if tc.expectedService.Name != serviceToCreate.Name {
					t.Errorf("expected %s, got %s", tc.expectedService.Name, serviceToCreate.Name)
				}
			}
		})
	}
}
