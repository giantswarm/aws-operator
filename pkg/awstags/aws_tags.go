package awstags

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/aws/aws-sdk-go/service/s3"
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

func NewKMS(tags map[string]string) []*kms.Tag {
	var ts []*kms.Tag
	for k, v := range tags {
		t := &kms.Tag{
			TagKey:   aws.String(k),
			TagValue: aws.String(v),
		}
		ts = append(ts, t)
	}

	return ts
}

func NewS3(tags map[string]string) []*s3.Tag {
	var ts []*s3.Tag
	for k, v := range tags {
		t := &s3.Tag{
			Key:   aws.String(k),
			Value: aws.String(v),
		}
		ts = append(ts, t)
	}

	return ts
}
