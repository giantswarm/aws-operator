package cloudformation

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	awscloudformation "github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/giantswarm/aws-operator/service/key"
	"github.com/giantswarm/microerror"
)

func (r *Resource) GetCurrentState(ctx context.Context, obj interface{}) (interface{}, error) {
	customObject, err := key.ToCustomObject(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	r.logger.Log("cluster", key.ClusterID(customObject), "debug", "looking for AWS stack")

	stackName := key.MainStackName(customObject)

	describeInput := &awscloudformation.DescribeStacksInput{
		StackName: aws.String(stackName),
	}
	describeOutput, err := r.awsClient.DescribeStacks(describeInput)

	// FIXME: The validation error returned by the CloudFormation API doesn't make things easy to check, other than
	// looking for the returned string. There's no constant in aws go sdk for defining this string, it comes from
	// the service.
	stackNotFoundError := fmt.Sprintf("Stack with id %s does not exist", stackName)
	if strings.Contains(err.Error(), stackNotFoundError) {

		r.logger.Log("cluster", key.ClusterID(customObject), "debug", notFoundError)
		return StackState{}, nil

	} else if err != nil {

		return nil, microerror.Mask(err)

	}

	if len(describeOutput.Stacks) > 1 {
		return nil, microerror.Mask(notFoundError)
	}

	outputStackState := StackState{
		Name:    stackName,
		Outputs: describeOutput.Stacks[0].Outputs,
	}

	return outputStackState, nil
}
