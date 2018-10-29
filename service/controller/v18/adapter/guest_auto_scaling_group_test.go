package adapter

import (
	"testing"

	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
)

func TestAdapterAutoScalingGroupRegularFields(t *testing.T) {
	t.Parallel()
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
			description: "having one worker leaves 1 worker head room for updates",
			customObject: v1alpha1.AWSConfig{
				Spec: v1alpha1.AWSConfigSpec{
					Cluster: defaultCluster,
					AWS: v1alpha1.AWSConfigSpecAWS{
						AZ: "myaz",
						Workers: []v1alpha1.AWSConfigSpecAWSNode{
							{},
						},
					},
				},
			},
			expectedError:                  false,
			expectedAZ:                     "myaz",
			expectedASGMaxSize:             2, // headroom is +1
			expectedASGMinSize:             1,
			expectedHealthCheckGracePeriod: gracePeriodSeconds,
			expectedMaxBatchSize:           "1",
			expectedMinInstancesInService:  "1",
			expectedRollingUpdatePauseTime: rollingUpdatePauseTime,
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
			expectedASGMaxSize:             4,
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
			expectedASGMaxSize:             8,
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
			err := a.Guest.AutoScalingGroup.Adapt(cfg)
			if tc.expectedError && err == nil {
				t.Error("expected error didn't happen")
			}

			if !tc.expectedError && err != nil {
				t.Errorf("unexpected error %v", err)
			}

			if !tc.expectedError {
				if a.Guest.AutoScalingGroup.ASGMaxSize != tc.expectedASGMaxSize {
					t.Errorf("unexpected output, got %d, want %d", a.Guest.AutoScalingGroup.ASGMaxSize, tc.expectedASGMaxSize)
				}

				if a.Guest.AutoScalingGroup.ASGMinSize != tc.expectedASGMinSize {
					t.Errorf("unexpected output, got %d, want %d", a.Guest.AutoScalingGroup.ASGMinSize, tc.expectedASGMinSize)
				}

				if a.Guest.AutoScalingGroup.HealthCheckGracePeriod != tc.expectedHealthCheckGracePeriod {
					t.Errorf("unexpected output, got %d, want %d", a.Guest.AutoScalingGroup.HealthCheckGracePeriod, tc.expectedHealthCheckGracePeriod)
				}

				if a.Guest.AutoScalingGroup.MaxBatchSize != tc.expectedMaxBatchSize {
					t.Errorf("unexpected output, got %q, want %q", a.Guest.AutoScalingGroup.MaxBatchSize, tc.expectedMaxBatchSize)
				}

				if a.Guest.AutoScalingGroup.MinInstancesInService != tc.expectedMinInstancesInService {
					t.Errorf("unexpected output, got %q, want %q", a.Guest.AutoScalingGroup.MinInstancesInService, tc.expectedMinInstancesInService)
				}

				if a.Guest.AutoScalingGroup.RollingUpdatePauseTime != tc.expectedRollingUpdatePauseTime {
					t.Errorf("unexpected output, got %q, want %q", a.Guest.AutoScalingGroup.RollingUpdatePauseTime, tc.expectedRollingUpdatePauseTime)
				}

				if a.Guest.AutoScalingGroup.WorkerAZ != tc.expectedAZ {
					t.Errorf("unexpected output, got %q, want %q", a.Guest.AutoScalingGroup.WorkerAZ, tc.expectedAZ)
				}

			}
		})
	}
}

func TestWorkerCountRatioMaxBatchSize(t *testing.T) {
	t.Parallel()
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
			description:   "scaling down to zero worker, ratio should be one, because it should never be zero",
			workers:       0,
			expectedRatio: "1",
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
