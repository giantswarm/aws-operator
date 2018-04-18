package cloudformation

import (
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	awscloudformation "github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
)

func Test_Resource_Cloudformation_GetCloudFormationTags(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		obj          v1alpha1.AWSConfig
		expectedTags []*awscloudformation.Tag
		description  string
	}{
		{
			description: "basic match",
			obj: v1alpha1.AWSConfig{
				Spec: v1alpha1.AWSConfigSpec{
					Cluster: v1alpha1.Cluster{
						ID: "5xchu",
					},
				},
			},
			expectedTags: []*awscloudformation.Tag{
				{
					Key:   aws.String("kubernetes.io/cluster/5xchu"),
					Value: aws.String("owned"),
				},
				{
					Key:   aws.String("giantswarm.io/cluster"),
					Value: aws.String("5xchu"),
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			tags := getCloudFormationTags(tc.obj)

			for _, tag := range tc.expectedTags {
				if !containsTag(tag, tags) {
					t.Fatalf("Expected cloud formation contains tag %v in the slice %v", tag, tags)
				}
			}
		})
	}
}

func containsTag(tag *awscloudformation.Tag, tags []*awscloudformation.Tag) bool {
	for _, inTag := range tags {
		if reflect.DeepEqual(tag, inTag) {
			return true
		}
	}

	return false
}
