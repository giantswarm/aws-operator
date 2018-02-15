package cloudformation

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	awscloudformation "github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
)

func Test_Resource_Cloudformation_GetCloudFormationTags(t *testing.T) {
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
				&awscloudformation.Tag{
					Key:   aws.String("kubernetes.io/cluster/5xchu"),
					Value: aws.String("owned"),
				},
				&awscloudformation.Tag{
					Key:   aws.String("KubernetesCluster"),
					Value: aws.String("5xchu"),
				},
			},
		},
	}

	noContainsTag := func(tag *awscloudformation.Tag, tags []*awscloudformation.Tag) bool {
		for _, inTag := range tags {
			fmt.Printf("%t %t \n", tag, inTag)
			if tag == inTag {
				return true
			}
		}

		return false
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			tags := getCloudFormationTags(tc.obj)

			for _, tag := range tc.expectedTags {
				if noContainsTag(tag, tags) {
					t.Fatalf("Expected cloud formation contains tag %v in the slice %v", tag, tags)
				}
			}
		})
	}
}
