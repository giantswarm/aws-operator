package adapter

import (
	"strconv"
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
			description: "basic matching, all fields present",
			customObject: v1alpha1.AWSConfig{
				Spec: v1alpha1.AWSConfigSpec{
					Cluster: defaultCluster,
					AWS: v1alpha1.AWSConfigSpecAWS{
						AZ: "myaz",
						Workers: []v1alpha1.AWSConfigSpecAWSNode{
							v1alpha1.AWSConfigSpecAWSNode{},
							v1alpha1.AWSConfigSpecAWSNode{},
							v1alpha1.AWSConfigSpecAWSNode{},
						},
					},
				},
			},
			expectedAZ:                     "myaz",
			expectedASGMaxSize:             3,
			expectedASGMinSize:             3,
			expectedHealthCheckGracePeriod: gracePeriodSeconds,
			expectedMaxBatchSize:           strconv.FormatFloat(asgMaxBatchSizeRatio, 'f', -1, 32),
			expectedMinInstancesInService:  strconv.FormatFloat(asgMinInstancesRatio, 'f', -1, 32),
			expectedRollingUpdatePauseTime: rollingUpdatePauseTime,
		},
	}

	for _, tc := range testCases {
		clients := Clients{}
		a := Adapter{}
		t.Run(tc.description, func(t *testing.T) {
			cfg := Config{
				CustomObject: tc.customObject,
				Clients:      clients,
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
