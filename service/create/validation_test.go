package create

import (
	"testing"

	"github.com/giantswarm/awstpr/spec/aws"
	"github.com/giantswarm/clustertpr/spec"
	"github.com/giantswarm/microerror"
	"github.com/stretchr/testify/assert"
)

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
