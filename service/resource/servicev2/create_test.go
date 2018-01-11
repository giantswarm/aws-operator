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
		Obj             interface{}
		Cur             interface{}
		Des             interface{}
		ExpectedService *apiv1.Service
	}{
		{
			Obj: &v1alpha1.AWSConfig{
				Spec: v1alpha1.AWSConfigSpec{
					Cluster: v1alpha1.Cluster{
						ID: "foobar",
					},
				},
			},
			Cur: &apiv1.Service{
				TypeMeta: apismetav1.TypeMeta{
					Kind:       "Service",
					APIVersion: "v1",
				},
				ObjectMeta: apismetav1.ObjectMeta{
					Name: "al9qy",
				},
			},
			Des: &apiv1.Service{
				TypeMeta: apismetav1.TypeMeta{
					Kind:       "Service",
					APIVersion: "v1",
				},
				ObjectMeta: apismetav1.ObjectMeta{
					Name: "al9qy",
				},
			},
			ExpectedService: nil,
		},

		{
			Obj: &v1alpha1.AWSConfig{
				Spec: v1alpha1.AWSConfigSpec{
					Cluster: v1alpha1.Cluster{
						ID: "foobar",
					},
				},
			},
			Cur: nil,
			Des: &apiv1.Service{
				TypeMeta: apismetav1.TypeMeta{
					Kind:       "Service",
					APIVersion: "v1",
				},
				ObjectMeta: apismetav1.ObjectMeta{
					Name: "al9qy",
				},
			},
			ExpectedService: &apiv1.Service{
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

	for i, tc := range testCases {
		result, err := newResource.newCreateChange(context.TODO(), tc.Obj, tc.Cur, tc.Des)
		if err != nil {
			t.Fatal("case", i+1, "expected", nil, "got", err)
		}
		if tc.ExpectedService == nil {
			if tc.ExpectedService != result {
				t.Fatal("case", i+1, "expected", tc.ExpectedService, "got", result)
			}
		} else {
			name := result.(*apiv1.Service).Name
			if tc.ExpectedService.Name != name {
				t.Fatal("case", i+1, "expected", tc.ExpectedService.Name, "got", name)
			}
		}
	}
}
