package adapter

import (
	"testing"

	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
)

func TestAdapterRouteTablesRegularFields(t *testing.T) {
	testCases := []struct {
		description                   string
		customObject                  v1alpha1.AWSConfig
		expectedError                 bool
		expectedPublicRouteTableName  string
		expectedPrivateRouteTableName string
	}{
		{
			description: "basic matching, all fields present",
			customObject: v1alpha1.AWSConfig{
				Spec: v1alpha1.AWSConfigSpec{
					Cluster: v1alpha1.Cluster{
						ID: "test-cluster",
					},
				},
			},
			expectedError:                 false,
			expectedPublicRouteTableName:  "test-cluster-public",
			expectedPrivateRouteTableName: "test-cluster-private",
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
			err := a.getRouteTables(cfg)
			if tc.expectedError && err == nil {
				t.Error("expected error didn't happen")
			}

			if !tc.expectedError && err != nil {
				t.Errorf("unexpected error %v", err)
			}

			if a.PublicRouteTableName != tc.expectedPublicRouteTableName {
				t.Errorf("unexpected PublicRouteTableName, got %q, want %q", a.PublicRouteTableName, tc.expectedPrivateRouteTableName)
			}

			if a.PrivateRouteTableName != tc.expectedPrivateRouteTableName {
				t.Errorf("unexpected PrivateRouteTableName, got %q, want %q", a.PrivateRouteTableName, tc.expectedPrivateRouteTableName)
			}
		})
	}
}
