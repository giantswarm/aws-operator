package adapter

import (
	"strconv"
	"testing"

	"github.com/giantswarm/awstpr"
	awsspec "github.com/giantswarm/awstpr/spec"
	awsspecaws "github.com/giantswarm/awstpr/spec/aws"
	"github.com/giantswarm/clustertpr"
	"github.com/giantswarm/clustertpr/spec"
)

func TestAdapterAutoScalingGroupRegularFields(t *testing.T) {
	testCases := []struct {
		description                    string
		customObject                   awstpr.CustomObject
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
			description:   "empty custom object",
			customObject:  awstpr.CustomObject{},
			expectedError: true,
		},
		{
			description: "basic matching, all fields present",
			customObject: awstpr.CustomObject{
				Spec: awstpr.Spec{
					Cluster: defaultCluster,
					AWS: awsspec.AWS{
						AZ: "myaz",
						Workers: []awsspecaws.Node{
							awsspecaws.Node{},
							awsspecaws.Node{},
							awsspecaws.Node{},
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

	clients := Clients{
		EC2: &EC2ClientMock{},
	}
	for _, tc := range testCases {
		a := Adapter{}
		t.Run(tc.description, func(t *testing.T) {
			err := a.getAutoScalingGroup(tc.customObject, clients)
			if tc.expectedError && err == nil {
				t.Error("expected error didn't happen")
			}

			if !tc.expectedError && err != nil {
				t.Errorf("unexpected error %v", err)
			}

			if !tc.expectedError {
				if a.autoScalingGroupAdapter.AZ != tc.expectedAZ {
					t.Errorf("unexpected output, got %q, want %q", a.autoScalingGroupAdapter.AZ, tc.expectedAZ)
				}

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
			}
		})
	}
}

func TestAdapterAutoScalingGroupLoadBalancerName(t *testing.T) {
	testCases := []struct {
		description              string
		customObject             awstpr.CustomObject
		expectedLoadBalancerName string
	}{
		{
			description: "basic matching, all fields present",
			customObject: awstpr.CustomObject{
				Spec: awstpr.Spec{
					Cluster: defaultCluster,
				},
			},
			expectedLoadBalancerName: "test-cluster-mysubdomain",
		},
	}

	clients := Clients{
		EC2: &EC2ClientMock{},
	}
	for _, tc := range testCases {
		a := Adapter{}
		t.Run(tc.description, func(t *testing.T) {
			err := a.getAutoScalingGroup(tc.customObject, clients)
			if err != nil {
				t.Errorf("unexpected error %v", err)
			}

			if a.LoadBalancerName != tc.expectedLoadBalancerName {
				t.Errorf("unexpected output, got %q, want %q", a.LoadBalancerName, tc.expectedLoadBalancerName)
			}
		})
	}
}

func TestAdapterAutoScalingGroupSubnetID(t *testing.T) {
	testCases := []struct {
		description                string
		customObject               awstpr.CustomObject
		expectedReceivedSubnetName string
		expectedError              bool
		unexistentSubnet           bool
	}{
		{
			description: "existent subnet",
			customObject: awstpr.CustomObject{
				Spec: awstpr.Spec{
					Cluster: defaultCluster,
					AWS: awsspec.AWS{
						Workers: []awsspecaws.Node{
							awsspecaws.Node{
								ImageID:      "myimageid",
								InstanceType: "myinstancetype",
							},
						},
					},
				},
			},
			expectedReceivedSubnetName: "test-cluster-public",
		},
		{
			description: "unexistent subnet",
			customObject: awstpr.CustomObject{
				Spec: awstpr.Spec{
					Cluster: clustertpr.Spec{
						Cluster: spec.Cluster{
							ID: "test-cluster",
						},
					},
					AWS: awsspec.AWS{
						Workers: []awsspecaws.Node{
							awsspecaws.Node{
								ImageID:      "myimageid",
								InstanceType: "myinstancetype",
							},
						},
					},
				},
			},
			unexistentSubnet:           true,
			expectedError:              true,
			expectedReceivedSubnetName: "",
		},
	}

	for _, tc := range testCases {
		a := Adapter{}
		clients := Clients{
			EC2: &EC2ClientMock{
				unexistingSubnet: tc.unexistentSubnet,
				clusterID:        "test-cluster",
			},
			IAM: &IAMClientMock{},
		}

		t.Run(tc.description, func(t *testing.T) {
			err := a.getAutoScalingGroup(tc.customObject, clients)
			if tc.expectedError && err == nil {
				t.Error("expected error didn't happen")
			}

			// the mock does the check internally, the returned subnet id is not related
			// to input
			if !tc.expectedError && err != nil {
				t.Errorf("unexpected error %v", err)
			}
		})
	}
}
