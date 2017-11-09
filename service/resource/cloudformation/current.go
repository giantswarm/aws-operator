package cloudformation

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	awscloudformation "github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/giantswarm/aws-operator/service/key"
	"github.com/giantswarm/microerror"
)

func (r *Resource) GetCurrentState(ctx context.Context, obj interface{}) (interface{}, error) {
	customObject, err := key.ToCustomObject(obj)
	if err != nil {
		return StackState{}, microerror.Mask(err)
	}

	r.logger.Log("cluster", key.ClusterID(customObject), "debug", "looking for AWS stack")

	stackName := key.MainStackName(customObject)

	describeInput := &awscloudformation.DescribeStacksInput{
		StackName: aws.String(stackName),
	}
	describeOutput, err := r.awsClient.DescribeStacks(describeInput)

	if IsStackNotFound(err) {

		r.logger.Log("cluster", key.ClusterID(customObject), "debug", "did not find a stack in AWS API")
		return StackState{}, nil

	} else if err != nil {

		return StackState{}, microerror.Mask(err)

	}

	if len(describeOutput.Stacks) > 1 {
		return StackState{}, microerror.Mask(notFoundError)
	}

	outputStackState := StackState{
		Name:    stackName,
		Outputs: describeOutput.Stacks[0].Outputs,
	}

	return outputStackState, nil
}
