package create

import (
	"testing"

	"github.com/giantswarm/awstpr/spec/aws"
	"github.com/juju/errgo"
	"github.com/stretchr/testify/assert"
)

func TestValidateWorkers(t *testing.T) {
	tests := []struct {
		name          string
		workers       []aws.Node
		expectedError error
	}{
		{
			name: "Valid workers - image IDs and instance types are the same",
			workers: []aws.Node{
				aws.Node{
					ImageID:      "example-image-id",
					InstanceType: "example-instance-type",
				},
				aws.Node{
					ImageID:      "example-image-id",
					InstanceType: "example-instance-type",
				},
			},
			expectedError: nil,
		},
		{
			name:          "Invalid workers - list is empty",
			workers:       []aws.Node{},
			expectedError: workersListEmptyError,
		},
		{
			name: "Invalid workers - image IDs are different",
			workers: []aws.Node{
				aws.Node{
					ImageID:      "example-image-id",
					InstanceType: "example-instance-type",
				},
				aws.Node{
					ImageID:      "another-image-id",
					InstanceType: "example-instance-type",
				},
			},
			expectedError: differentImageIDsError,
		},
		{
			name: "Invalid workers - instance types are different",
			workers: []aws.Node{
				aws.Node{
					ImageID:      "example-image-id",
					InstanceType: "example-instance-type",
				},
				aws.Node{
					ImageID:      "example-image-id",
					InstanceType: "another-instance-type",
				},
			},
			expectedError: differentInstanceTypesError,
		},
	}

	for _, tc := range tests {
		err := validateWorkers(tc.workers)
		assert.Equal(t, tc.expectedError, errgo.Cause(err), tc.name)
	}
}
