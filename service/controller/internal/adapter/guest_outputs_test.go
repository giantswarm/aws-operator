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
		ExpectedMasterIgnitionHash       string
	}{
		{
			Description: "master CloudConfig version should match the hardcoded value",
			Config: Config{
				CustomObject: v1alpha1.AWSConfig{},
				StackState: StackState{
					MasterIgnitionHash: "foo",
				},
			},
			ExpectedMasterIgnitionHash: "foo",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Description, func(t *testing.T) {
			a := &GuestOutputsAdapter{}

			err := a.Adapt(tc.Config)
			if err != nil {
				t.Fatalf("expected %#v got %#v", nil, err)
			}

			if a.Master.CloudConfig.Hash != tc.ExpectedMasterIgnitionHash {
				t.Fatalf("expected %s got %s", tc.ExpectedMasterIgnitionHash, a.Master.CloudConfig.Hash)
			}
		})
	}
}

func Test_CloudFormation_Adapter_Outputs_WorkerCloudConfigVersion(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		Description                      string
		Config                           Config
		ExpectedWorkerIgnitionHash       string
	}{
		{
			Description: "worker CloudConfig version should match the hardcoded value",
			Config: Config{
				CustomObject: v1alpha1.AWSConfig{},
				StackState: StackState{
					WorkerIgnitionHash: "foo",
				},
			},
			ExpectedWorkerIgnitionHash: "foo",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Description, func(t *testing.T) {
			a := &GuestOutputsAdapter{}

			err := a.Adapt(tc.Config)
			if err != nil {
				t.Fatalf("expected %#v got %#v", nil, err)
			}

			if a.Worker.CloudConfig.Hash != tc.ExpectedWorkerIgnitionHash {
				t.Fatalf("expected %s got %s", tc.ExpectedWorkerIgnitionHash, a.Worker.CloudConfig.Hash)
			}
		})
	}
}
