package adapter

import (
	"testing"

	"github.com/giantswarm/aws-operator/service/cloudconfigv3"
)

func TestAdapterOutputsRegularFields(t *testing.T) {
	testCases := []struct {
		description                      string
		expectedMasterCloudConfigVersion string
		expectedWorkerCloudConfigVersion string
	}{
		{
			description:                      "basic check",
			expectedMasterCloudConfigVersion: cloudconfigv3.MasterCloudConfigVersion,
			expectedWorkerCloudConfigVersion: cloudconfigv3.WorkerCloudConfigVersion,
		},
	}
	for _, tc := range testCases {
		a := Adapter{}
		clients := Clients{}
		t.Run(tc.description, func(t *testing.T) {
			cfg := Config{
				Clients: clients,
			}
			err := a.getOutputs(cfg)

			if err != nil {
				t.Errorf("unexpected error %v", err)
			}

			if a.MasterCloudConfigVersion != tc.expectedMasterCloudConfigVersion {
				t.Errorf("unexpected MasterCloudConfigVersion, got %q, want %q", a.MasterCloudConfigVersion, tc.expectedMasterCloudConfigVersion)
			}

			if a.WorkerCloudConfigVersion != tc.expectedWorkerCloudConfigVersion {
				t.Errorf("unexpected WorkerCloudConfigVersion, got %q, want %q", a.WorkerCloudConfigVersion, tc.expectedWorkerCloudConfigVersion)
			}
		})
	}
}
