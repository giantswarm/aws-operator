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
				Clients:      Clients{},
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
				Clients:      Clients{},
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

func Test_CloudFormation_Adapter_Outputs_WorkerCount(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		Description       string
		Config            Config
		ExpectedWorkerMax int
		ExpectedWorkerMin int
	}{
		{
			Description: "worker's max/min value should match the configuration",
			Config: Config{
				Clients:      Clients{},
				CustomObject: v1alpha1.AWSConfig{},
				StackState: StackState{
					WorkerMax: 10,
					WorkerMin: 3,
				},
			},
			ExpectedWorkerMax: 10,
			ExpectedWorkerMin: 3,
		},

		{
			Description: "worker' max/min value should match the configuration",
			Config: Config{
				Clients:      Clients{},
				CustomObject: v1alpha1.AWSConfig{},
				StackState: StackState{
					WorkerMax: 3,
					WorkerMin: 3,
				},
			},
			ExpectedWorkerMax: 3,
			ExpectedWorkerMin: 3,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Description, func(t *testing.T) {
			a := &GuestOutputsAdapter{}

			err := a.Adapt(tc.Config)
			if err != nil {
				t.Fatalf("expected %#v got %#v", nil, err)
			}

			if a.Worker.Max != tc.ExpectedWorkerMax {
				t.Fatalf("expected max: %d got %d", tc.ExpectedWorkerMax, a.Worker.Max)
			}
			if a.Worker.Min != tc.ExpectedWorkerMin {
				t.Fatalf("expected min: %d got %d", tc.ExpectedWorkerMin, a.Worker.Min)
			}
		})
	}
}
