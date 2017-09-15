package aws

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/autoscaling"
)

func TestIsLegacyASG(t *testing.T) {
	nameOk := "my-asg"
	clusterOk := "my-cluster"
	tagNameOK := tagKeyName
	tagClusterOK := tagKeyCluster
	tagCF := "aws:cloudformation:stack-id"
	nameBad := "my-other-asg"
	clusterBad := "my-other-cluster"

	testCases := []struct {
		description string
		name        string
		clusterID   string
		tags        []*autoscaling.TagDescription
		expected    bool
	}{
		{
			description: "name and cluster match, no CF tag",
			name:        nameOk,
			clusterID:   clusterOk,
			tags: []*autoscaling.TagDescription{
				{Key: &tagNameOK, Value: &nameOk},
				{Key: &tagClusterOK, Value: &clusterOk},
			},
			expected: true},
		{
			description: "name and cluster match, CF tag",
			name:        nameOk,
			clusterID:   clusterOk,
			tags: []*autoscaling.TagDescription{
				{Key: &tagNameOK, Value: &nameOk},
				{Key: &tagClusterOK, Value: &clusterOk},
				{Key: &tagCF, Value: &clusterOk},
			},
			expected: false},
		{
			description: "name no match, cluster match, no CF tag",
			name:        nameBad,
			clusterID:   clusterOk,
			tags: []*autoscaling.TagDescription{
				{Key: &tagNameOK, Value: &nameOk},
				{Key: &tagClusterOK, Value: &clusterOk},
			},
			expected: false},
		{
			description: "name match, cluster no match, no CF tag",
			name:        nameOk,
			clusterID:   clusterBad,
			tags: []*autoscaling.TagDescription{
				{Key: &tagNameOK, Value: &nameOk},
				{Key: &tagClusterOK, Value: &clusterOk},
			}, expected: false},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			actual := isLegacyASG(tc.name, tc.clusterID, tc.tags)

			if actual != tc.expected {
				t.Errorf("got %v, expected %v for tags %v, name %s and clusterID %s",
					actual, tc.expected, tc.tags, tc.name, tc.clusterID)
			}
		})
	}
}
