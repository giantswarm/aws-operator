package namespace

import (
	"reflect"
	"testing"

	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
)

func Test_getNamespace(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		description    string
		obj            v1alpha1.AWSConfig
		expectedName   string
		expectedLabels map[string]string
	}{
		{
			description: "case 0: basic match",
			obj: v1alpha1.AWSConfig{
				Spec: v1alpha1.AWSConfigSpec{
					Cluster: v1alpha1.Cluster{
						ID: "al9qy",
						Customer: v1alpha1.ClusterCustomer{
							ID: "giantswarm",
						},
					},
				},
			},
			expectedName: "al9qy",
			expectedLabels: map[string]string{
				"cluster":  "al9qy",
				"customer": "giantswarm",
			},
		},
		{
			description: "case 1: different cluster id",
			obj: v1alpha1.AWSConfig{
				Spec: v1alpha1.AWSConfigSpec{
					Cluster: v1alpha1.Cluster{
						ID: "foobar",
						Customer: v1alpha1.ClusterCustomer{
							ID: "giantswarm",
						},
					},
				},
			},
			expectedName: "foobar",
			expectedLabels: map[string]string{
				"cluster":  "foobar",
				"customer": "giantswarm",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			result := getNamespace(tc.obj)

			if tc.expectedName != result.Name {
				t.Fatalf("expected name %s got %s", tc.expectedName, result.Name)
			}

			if !reflect.DeepEqual(tc.expectedLabels, result.Labels) {
				t.Fatalf("expected labels %q got %q", tc.expectedLabels, result.Labels)
			}
		})
	}
}
