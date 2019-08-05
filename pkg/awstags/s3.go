package awstags

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
)

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
