package create

import (
	"testing"

	"github.com/giantswarm/awstpr"

	awsspec "github.com/giantswarm/awstpr/spec"
	"github.com/giantswarm/awstpr/spec/aws"
	"github.com/giantswarm/clustertpr/spec"
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
		cluster := awstpr.CustomObject{
			Spec: awstpr.Spec{
				AWS: awsspec.AWS{
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
		awsWorkers    []aws.Node
		workers       []spec.Node
		expectedError error
	}{
		{
			name: "Valid workers - image IDs and instance types are the same",
			awsWorkers: []aws.Node{
				{
					ImageID:      "example-image-id",
					InstanceType: "example-instance-type",
				},
				{
					ImageID:      "example-image-id",
					InstanceType: "example-instance-type",
				},
			},
			workers: []spec.Node{
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
			awsWorkers:    []aws.Node{},
			workers:       []spec.Node{},
			expectedError: workersListEmptyError,
		},
		{
			name:       "Invalid workers - aws workers list is empty",
			awsWorkers: []aws.Node{},
			workers: []spec.Node{
				{
					ID: "worker-1",
				},
			},
			expectedError: workersListEmptyError,
		},
		{
			name: "Invalid workers - image IDs are different",
			awsWorkers: []aws.Node{
				{
					ImageID:      "example-image-id",
					InstanceType: "example-instance-type",
				},
				{
					ImageID:      "another-image-id",
					InstanceType: "example-instance-type",
				},
			},
			workers: []spec.Node{
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
			awsWorkers: []aws.Node{
				{
					ImageID:      "example-image-id",
					InstanceType: "example-instance-type",
				},
				{
					ImageID:      "example-image-id",
					InstanceType: "another-instance-type",
				},
			},
			workers: []spec.Node{
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
			awsWorkers: []aws.Node{
				{
					ImageID:      "example-image-id",
					InstanceType: "example-instance-type",
				},
				{
					ImageID:      "example-image-id",
					InstanceType: "another-instance-type",
				},
			},
			workers: []spec.Node{
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
