package cloudformation

import (
	"github.com/aws/aws-sdk-go/aws"
	awsCF "github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/giantswarm/aws-operator/service/key"
	"github.com/giantswarm/awstpr"
)

func newMainStack(customObject awstpr.CustomObject) (*awsCF.CreateStackInput, error) {
	stackName := key.MainStackName(customObject)

	mainCF := &awsCF.CreateStackInput{
		StackName: aws.String(stackName),
	}

	return mainCF, nil
}
