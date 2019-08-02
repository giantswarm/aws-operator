package awstags

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
)

func NewCloudFormation(tags map[string]string) []*cloudformation.Tag {
	var ts []*cloudformation.Tag
	for k, v := range tags {
		t := &cloudformation.Tag{
			Key:   aws.String(k),
			Value: aws.String(v),
		}
		ts = append(ts, t)
	}

	return ts
}
