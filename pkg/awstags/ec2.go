package awstags

import "github.com/aws/aws-sdk-go/service/ec2"

func HasTag(tags []*ec2.Tag, key string) bool {
	for _, t := range tags {
		if *t.Key == key {
			return true
		}
	}

	return false
}

func HasTags(tags []*ec2.Tag, keys ...string) bool {
	for _, k := range keys {
		if !HasTag(tags, k) {
			return false
		}
	}

	return true
}

func ValueForKey(tags []*ec2.Tag, key string) string {
	for _, t := range tags {
		if *t.Key == key {
			return *t.Value
		}
	}

	return ""
}
