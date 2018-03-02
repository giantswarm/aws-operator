package adapter

import (
	"testing"

	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
)

func TestAdapterAutoScalingGroupRegularFields(t *testing.T) {
	testCases := []struct {
		description                    string
		customObject                   v1alpha1.AWSConfig
		expectedError                  bool
		expectedAZ                     string
		expectedASGMaxSize             int
		expectedASGMinSize             int
		expectedHealthCheckGracePeriod int
		expectedMaxBatchSize           string
		expectedMinInstancesInService  string
		expectedRollingUpdatePauseTime string
	}{
		{
			description: "no worker nodes, return error",
			customObject: v1alpha1.AWSConfig{
				Spec: v1alpha1.AWSConfigSpec{
					Cluster: defaultCluster,
					AWS: v1alpha1.AWSConfigSpecAWS{
						AZ:      "myaz",
						Workers: []v1alpha1.AWSConfigSpecAWSNode{},
					},
				},
			},
			expectedError: true,
		},
		{
			description: "basic matching, all fields present",
			customObject: v1alpha1.AWSConfig{
				Spec: v1alpha1.AWSConfigSpec{
					Cluster: defaultCluster,
					AWS: v1alpha1.AWSConfigSpecAWS{
						AZ: "myaz",
						Workers: []v1alpha1.AWSConfigSpecAWSNode{
							{},
							{},
							{},
						},
					},
				},
			},
			expectedError:                  false,
			expectedAZ:                     "myaz",
			expectedASGMaxSize:             3,
			expectedASGMinSize:             3,
			expectedHealthCheckGracePeriod: gracePeriodSeconds,
			expectedMaxBatchSize:           "1",
			expectedMinInstancesInService:  "2",
			expectedRollingUpdatePauseTime: rollingUpdatePauseTime,
		},
		{
			description: "7 node cluster, batch size and min instances are correct",
			customObject: v1alpha1.AWSConfig{
				Spec: v1alpha1.AWSConfigSpec{
					Cluster: defaultCluster,
					AWS: v1alpha1.AWSConfigSpecAWS{
						AZ: "myaz",
						Workers: []v1alpha1.AWSConfigSpecAWSNode{
							{},
							{},
							{},
							{},
							{},
							{},
							{},
						},
					},
				},
			},
			expectedError:                  false,
			expectedAZ:                     "myaz",
			expectedASGMaxSize:             7,
			expectedASGMinSize:             7,
			expectedHealthCheckGracePeriod: gracePeriodSeconds,
			expectedMaxBatchSize:           "2",
			expectedMinInstancesInService:  "5",
			expectedRollingUpdatePauseTime: rollingUpdatePauseTime,
		},
	}

	for _, tc := range testCases {
		a := Adapter{}
		t.Run(tc.description, func(t *testing.T) {
			cfg := Config{
				CustomObject: tc.customObject,
				Clients:      Clients{},
				HostClients:  Clients{},
			}
			err := a.getAutoScalingGroup(cfg)
			if tc.expectedError && err == nil {
				t.Error("expected error didn't happen")
			}

			if !tc.expectedError && err != nil {
				t.Errorf("unexpected error %v", err)
			}

			if !tc.expectedError {
				if a.ASGMaxSize != tc.expectedASGMaxSize {
					t.Errorf("unexpected output, got %d, want %d", a.ASGMaxSize, tc.expectedASGMaxSize)
				}

				if a.ASGMinSize != tc.expectedASGMinSize {
					t.Errorf("unexpected output, got %d, want %d", a.ASGMinSize, tc.expectedASGMinSize)
				}

				if a.HealthCheckGracePeriod != tc.expectedHealthCheckGracePeriod {
					t.Errorf("unexpected output, got %d, want %d", a.HealthCheckGracePeriod, tc.expectedHealthCheckGracePeriod)
				}

				if a.MaxBatchSize != tc.expectedMaxBatchSize {
					t.Errorf("unexpected output, got %q, want %q", a.MaxBatchSize, tc.expectedMaxBatchSize)
				}

				if a.MinInstancesInService != tc.expectedMinInstancesInService {
					t.Errorf("unexpected output, got %q, want %q", a.MinInstancesInService, tc.expectedMinInstancesInService)
				}

				if a.RollingUpdatePauseTime != tc.expectedRollingUpdatePauseTime {
					t.Errorf("unexpected output, got %q, want %q", a.RollingUpdatePauseTime, tc.expectedRollingUpdatePauseTime)
				}

				if a.WorkerAZ != tc.expectedAZ {
					t.Errorf("unexpected output, got %q, want %q", a.WorkerAZ, tc.expectedAZ)
				}

			}
		})
	}
}

func TestWorkerCountRatioMaxBatchSize(t *testing.T) {
	tcs := []struct {
		description   string
		workers       int
		expectedRatio string
	}{
		{
			description:   "scaling down to one worker, one should be up",
			workers:       1,
			expectedRatio: "1",
		},
		{
			description:   "scaling down to two worker, one should be up",
			workers:       2,
			expectedRatio: "1",
		},
		{
			description:   "scaling down to zero worker, none should be up",
			workers:       0,
			expectedRatio: "0",
		},
	}
	for _, tc := range tcs {
		t.Run(tc.description, func(t *testing.T) {
			actual := workerCountRatio(tc.workers, asgMaxBatchSizeRatio)

			if tc.expectedRatio != actual {
				t.Errorf("wrong worker count ratio, expected %q, actual %q", tc.expectedRatio, actual)
			}
		})
	}
}
