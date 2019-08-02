package awstags

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kms"
)

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
