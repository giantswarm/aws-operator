package legacy

import (
	"testing"

	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/microerror"
	"github.com/stretchr/testify/assert"
)

func TestValidateAvailabilityZone(t *testing.T) {
	tests := []struct {
		name             string
		region           string
		availabilityZone string
		expectedError    error
	}{
		{
			name:             "Valid AZ",
			region:           "eu-central-1",
			availabilityZone: "eu-central-1a",
			expectedError:    nil,
		},
		{
			name:             "Valid AZ in a different region",
			region:           "us-east-1",
			availabilityZone: "us-east-1d",
			expectedError:    nil,
		},
		{
			name:             "Invalid AZ for region",
			region:           "eu-central-1",
			availabilityZone: "eu-west-1a",
			expectedError:    invalidAvailabilityZoneError,
		},
		{
			name:             "Invalid AZ format",
			region:           "eu-central-1",
			availabilityZone: "eu-central1a",
			expectedError:    invalidAvailabilityZoneError,
		},
		{
			name:             "Invalid numeric AZ",
			region:           "eu-central-1",
			availabilityZone: "eu-central-11",
			expectedError:    invalidAvailabilityZoneError,
		},
	}

	for _, tc := range tests {
		cluster := v1alpha1.AWSConfig{
			Spec: v1alpha1.AWSConfigSpec{
				AWS: v1alpha1.AWSConfigSpecAWS{
					Region: tc.region,
					AZ:     tc.availabilityZone,
				},
			},
		}

		err := validateAvailabilityZone(cluster)
		assert.Equal(t, tc.expectedError, microerror.Cause(err), tc.name)
	}
}

func TestValidateWorkers(t *testing.T) {
	tests := []struct {
		name          string
		awsWorkers    []v1alpha1.AWSConfigSpecAWSNode
		workers       []v1alpha1.ClusterNode
		expectedError error
	}{
		{
			name: "Valid workers - image IDs and instance types are the same",
			awsWorkers: []v1alpha1.AWSConfigSpecAWSNode{
				{
					ImageID:      "example-image-id",
					InstanceType: "example-instance-type",
				},
				{
					ImageID:      "example-image-id",
					InstanceType: "example-instance-type",
				},
			},
			workers: []v1alpha1.ClusterNode{
				{
					ID: "worker-1",
				},
				{
					ID: "worker-2",
				},
			},
			expectedError: nil,
		},
		{
			name:          "Invalid workers - list is empty",
			awsWorkers:    []v1alpha1.AWSConfigSpecAWSNode{},
			workers:       []v1alpha1.ClusterNode{},
			expectedError: workersListEmptyError,
		},
		{
			name:       "Invalid workers - aws workers list is empty",
			awsWorkers: []v1alpha1.AWSConfigSpecAWSNode{},
			workers: []v1alpha1.ClusterNode{
				{
					ID: "worker-1",
				},
			},
			expectedError: workersListEmptyError,
		},
		{
			name: "Invalid workers - image IDs are different",
			awsWorkers: []v1alpha1.AWSConfigSpecAWSNode{
				{
					ImageID:      "example-image-id",
					InstanceType: "example-instance-type",
				},
				{
					ImageID:      "another-image-id",
					InstanceType: "example-instance-type",
				},
			},
			workers: []v1alpha1.ClusterNode{
				{
					ID: "worker-1",
				},
				{
					ID: "worker-2",
				},
			},
			expectedError: differentImageIDsError,
		},
		{
			name: "Invalid workers - instance types are different",
			awsWorkers: []v1alpha1.AWSConfigSpecAWSNode{
				{
					ImageID:      "example-image-id",
					InstanceType: "example-instance-type",
				},
				{
					ImageID:      "example-image-id",
					InstanceType: "another-instance-type",
				},
			},
			workers: []v1alpha1.ClusterNode{
				{
					ID: "worker-1",
				},
				{
					ID: "worker-2",
				},
			},
			expectedError: differentInstanceTypesError,
		},
		{
			name: "Invalid workers - node counts are different",
			awsWorkers: []v1alpha1.AWSConfigSpecAWSNode{
				{
					ImageID:      "example-image-id",
					InstanceType: "example-instance-type",
				},
				{
					ImageID:      "example-image-id",
					InstanceType: "another-instance-type",
				},
			},
			workers: []v1alpha1.ClusterNode{
				{
					ID: "worker-1",
				},
			},
			expectedError: invalidWorkerNodeCountError,
		},
	}

	for _, tc := range tests {
		err := validateWorkers(tc.awsWorkers, tc.workers)
		assert.Equal(t, tc.expectedError, microerror.Cause(err), tc.name)
	}
}

func TestValidateELB(t *testing.T) {
	tests := []struct {
		name                      string
		idleTimeoutSecondsAPI     int
		idleTimeoutSecondsEtcd    int
		idleTimeoutSecondsIngress int
		expectedError             error
	}{
		{
			name: "Valid timeout",
			idleTimeoutSecondsAPI:     60,
			idleTimeoutSecondsEtcd:    60,
			idleTimeoutSecondsIngress: 60,
			expectedError:             nil,
		},
		{
			name: "Valid timeout negative timeout invokes default",
			idleTimeoutSecondsAPI:     -1,
			idleTimeoutSecondsEtcd:    -1,
			idleTimeoutSecondsIngress: -1,
			expectedError:             nil,
		},
		{
			name: "Valid timeout zero invokes default",
			idleTimeoutSecondsAPI:     0,
			idleTimeoutSecondsEtcd:    0,
			idleTimeoutSecondsIngress: 0,
			expectedError:             nil,
		},
		{
			name: "Invalid timeout exceeds maximum",
			idleTimeoutSecondsAPI:     3601,
			idleTimeoutSecondsEtcd:    3601,
			idleTimeoutSecondsIngress: 3601,
			expectedError:             idleTimeoutSecondsOutOfRangeError,
		},
	}

	for _, tc := range tests {
		aws := v1alpha1.AWSConfigSpecAWS{
			API: v1alpha1.AWSConfigSpecAWSAPI{
				ELB: v1alpha1.AWSConfigSpecAWSAPIELB{
					IdleTimeoutSeconds: tc.idleTimeoutSecondsAPI,
				},
			},
			Etcd: v1alpha1.AWSConfigSpecAWSEtcd{
				ELB: v1alpha1.AWSConfigSpecAWSEtcdELB{
					IdleTimeoutSeconds: tc.idleTimeoutSecondsEtcd,
				},
			},
			Ingress: v1alpha1.AWSConfigSpecAWSIngress{
				ELB: v1alpha1.AWSConfigSpecAWSIngressELB{
					IdleTimeoutSeconds: tc.idleTimeoutSecondsIngress,
				},
			},
		}

		err := validateELB(aws)
		assert.Equal(t, tc.expectedError, microerror.Cause(err), tc.name)
	}
}

// Specific test with a missing value in the ELB struct
func TestValidateELBSparseStruct(t *testing.T) {
	tests := []struct {
		name                   string
		idleTimeoutSecondsAPI  int
		idleTimeoutSecondsEtcd int
		expectedError          error
	}{
		{
			name: "Valid timeout missing value invokes default",
			idleTimeoutSecondsAPI:  60,
			idleTimeoutSecondsEtcd: 60,
			expectedError:          nil,
		},
	}

	for _, tc := range tests {
		aws := v1alpha1.AWSConfigSpecAWS{
			API: v1alpha1.AWSConfigSpecAWSAPI{
				ELB: v1alpha1.AWSConfigSpecAWSAPIELB{
					IdleTimeoutSeconds: tc.idleTimeoutSecondsAPI,
				},
			},
			Etcd: v1alpha1.AWSConfigSpecAWSEtcd{
				ELB: v1alpha1.AWSConfigSpecAWSEtcdELB{
					IdleTimeoutSeconds: tc.idleTimeoutSecondsEtcd,
				},
			},
			// Ingress:
		}

		err := validateELB(aws)
		assert.Equal(t, tc.expectedError, microerror.Cause(err), tc.name)
	}
}

// Specific test with a missing IdleTimeoutSeconds struct
func TestValidateELBMissingStruct(t *testing.T) {
	tests := []struct {
		name          string
		expectedError error
	}{
		{
			name:          "Valid timeout missing value invokes default",
			expectedError: nil,
		},
	}

	for _, tc := range tests {
		// Missing nested structs
		aws := v1alpha1.AWSConfigSpecAWS{}

		err := validateELB(aws)
		assert.Equal(t, tc.expectedError, microerror.Cause(err), tc.name)
	}
}
