package aws

import (
	microerror "github.com/giantswarm/microkit/error"

	"github.com/aws/aws-sdk-go/service/sns"
)

type Topic struct {
	Name string
	AWSEntity
}

func (t Topic) findExisting() (*sns.Topic, error) {
	topics, err := t.Clients.SNS.ListTopics(&sns.ListTopicsInput{})
	if err != nil {
		return nil, microerror.MaskAny(err)
	}

	if len(topics.Topics) < 1 {
		return nil, microerror.MaskAnyf(notFoundError, notFoundErrorFormat, TopicType, t.Name)
	} else if len(topics.Topics) > 1 {
		return nil, microerror.MaskAny(tooManyResultsError)
	}

	return topics.Topics[0], nil
}
