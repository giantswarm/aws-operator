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
		{"name and cluster match, no CF tag", nameOk, clusterOk, []*autoscaling.TagDescription{
			{Key: &tagNameOK, Value: &nameOk},
			{Key: &tagClusterOK, Value: &clusterOk},
		}, true},
		{"name and cluster match, CF tag", nameOk, clusterOk, []*autoscaling.TagDescription{
			{Key: &tagNameOK, Value: &nameOk},
			{Key: &tagClusterOK, Value: &clusterOk},
			{Key: &tagCF, Value: &clusterOk},
		}, false},
		{"name no match, cluster match, no CF tag", nameBad, clusterOk, []*autoscaling.TagDescription{
			{Key: &tagNameOK, Value: &nameOk},
			{Key: &tagClusterOK, Value: &clusterOk},
		}, false},
		{"name match, cluster no match, no CF tag", nameOk, clusterBad, []*autoscaling.TagDescription{
			{Key: &tagNameOK, Value: &nameOk},
			{Key: &tagClusterOK, Value: &clusterOk},
		}, false},
	}

	for _, tc := range testCases {
		t.Run("name and cluster match, no CF tag", func(t *testing.T) {
			actual := isLegacyASG(tc.name, tc.clusterID, tc.tags)

			if actual != tc.expected {
				t.Errorf("got %v, expected %v for tags %v, name %s and clusterID %s",
					actual, tc.expected, tc.tags, tc.name, tc.clusterID)
			}
		})
	}
}
