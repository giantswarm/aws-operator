package adapter

import (
	"testing"

	"github.com/giantswarm/awstpr"
	"github.com/giantswarm/clustertpr"
)

func TestAdapterOutputsRegularFields(t *testing.T) {
	testCases := []struct {
		description            string
		customObject           awstpr.CustomObject
		expectedClusterVersion string
	}{
		{
			description:            "empty custom object",
			customObject:           awstpr.CustomObject{},
			expectedClusterVersion: "",
		},
		{
			description: "basic matching",
			customObject: awstpr.CustomObject{
				Spec: awstpr.Spec{
					Cluster: clustertpr.Spec{
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
			err := a.getOutputs(tc.customObject, clients)

			if err != nil {
				t.Errorf("unexpected error %v", err)
			}

			if a.ClusterVersion != tc.expectedClusterVersion {
				t.Errorf("unexpected ClusterVersion, got %q, want %q", a.ClusterVersion, tc.expectedClusterVersion)
			}
		})
	}
}
