package adapter

import (
	"testing"

	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
)

func Test_CloudFormation_Adapter_Outputs_MasterCloudConfigVersion(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		Description                      string
		Config                           Config
		ExpectedMasterCloudConfigVersion string
	}{
		{
			Description: "master CloudConfig version should match the hardcoded value",
			Config: Config{
				CustomObject: v1alpha1.AWSConfig{},
				StackState: StackState{
					MasterCloudConfigVersion: "foo",
				},
			},
			ExpectedMasterCloudConfigVersion: "foo",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Description, func(t *testing.T) {
			a := &GuestOutputsAdapter{}

			err := a.Adapt(tc.Config)
			if err != nil {
				t.Fatalf("expected %#v got %#v", nil, err)
			}

			if a.Master.CloudConfig.Version != tc.ExpectedMasterCloudConfigVersion {
				t.Fatalf("expected %s got %s", tc.ExpectedMasterCloudConfigVersion, a.Master.CloudConfig.Version)
			}
		})
	}
}

func Test_CloudFormation_Adapter_Outputs_WorkerCloudConfigVersion(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		Description                      string
		Config                           Config
		ExpectedWorkerCloudConfigVersion string
	}{
		{
			Description: "worker CloudConfig version should match the hardcoded value",
			Config: Config{
				CustomObject: v1alpha1.AWSConfig{},
				StackState: StackState{
					WorkerCloudConfigVersion: "foo",
				},
			},
			ExpectedWorkerCloudConfigVersion: "foo",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Description, func(t *testing.T) {
			a := &GuestOutputsAdapter{}

			err := a.Adapt(tc.Config)
			if err != nil {
				t.Fatalf("expected %#v got %#v", nil, err)
			}

			if a.Worker.CloudConfig.Version != tc.ExpectedWorkerCloudConfigVersion {
				t.Fatalf("expected %s got %s", tc.ExpectedWorkerCloudConfigVersion, a.Worker.CloudConfig.Version)
			}
		})
	}
}
