package adapter

import (
	"testing"

	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
)

func TestAdapterOutputsRegularFields(t *testing.T) {
	testCases := []struct {
		description            string
		customObject           v1alpha1.AWSConfig
		expectedClusterVersion string
	}{
		{
			description:            "empty custom object",
			customObject:           v1alpha1.AWSConfig{},
			expectedClusterVersion: "",
		},
		{
			description: "basic matching",
			customObject: v1alpha1.AWSConfig{
				Spec: v1alpha1.AWSConfigSpec{
					Cluster: v1alpha1.Cluster{
						Version: "myversion",
					},
				},
			},
			expectedClusterVersion: "myversion",
		},
	}
	for _, tc := range testCases {
		a := Adapter{}
		clients := Clients{}
		t.Run(tc.description, func(t *testing.T) {
			cfg := Config{
				CustomObject: tc.customObject,
				Clients:      clients,
			}
			err := a.getOutputs(cfg)

			if err != nil {
				t.Errorf("unexpected error %v", err)
			}

			if a.ClusterVersion != tc.expectedClusterVersion {
				t.Errorf("unexpected ClusterVersion, got %q, want %q", a.ClusterVersion, tc.expectedClusterVersion)
			}
		})
	}
}
