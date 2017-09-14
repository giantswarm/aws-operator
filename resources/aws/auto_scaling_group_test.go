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

	t.Run("name and cluster match, no CF tag", func(t *testing.T) {
		tags := []*autoscaling.TagDescription{
			{Key: &tagNameOK, Value: &nameOk},
			{Key: &tagClusterOK, Value: &clusterOk},
		}

		if !isLegacyASG(nameOk, clusterOk, tags) {
			t.Errorf("not legacy for tags %v, name %s and clusterID %s",
				tags, nameOk, clusterOk)
		}
	})

	t.Run("name and cluster match, CF tag", func(t *testing.T) {
		tags := []*autoscaling.TagDescription{
			{Key: &tagNameOK, Value: &nameOk},
			{Key: &tagClusterOK, Value: &clusterOk},
			{Key: &tagCF, Value: &clusterOk},
		}

		if isLegacyASG(nameOk, clusterOk, tags) {
			t.Errorf("legacy for tags %v, name %s and clusterID %s",
				tags, nameOk, clusterOk)
		}
	})

	t.Run("name no match, cluster match, no CF tag", func(t *testing.T) {
		tags := []*autoscaling.TagDescription{
			{Key: &tagNameOK, Value: &nameOk},
			{Key: &tagClusterOK, Value: &clusterOk},
		}

		nameBad := "my-other-asg"

		if isLegacyASG(nameBad, clusterOk, tags) {
			t.Errorf("legacy for tags %v, name %s and clusterID %s",
				tags, nameBad, clusterOk)
		}
	})

	t.Run("name match, cluster no match, no CF tag", func(t *testing.T) {
		tags := []*autoscaling.TagDescription{
			{Key: &tagNameOK, Value: &nameOk},
			{Key: &tagClusterOK, Value: &clusterOk},
		}

		clusterBad := "my-other-cluster"

		if isLegacyASG(nameOk, clusterBad, tags) {
			t.Errorf("legacy for tags %v, name %s and clusterID %s",
				tags, clusterBad, clusterBad)
		}
	})

}
