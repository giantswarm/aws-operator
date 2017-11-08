package cloudformation

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	awsCF "github.com/aws/aws-sdk-go/service/cloudformation"
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

	describeInput := &awsCF.DescribeStacksInput{
		StackName: aws.String(stackName),
	}
	describeOutput, err := r.awsClient.DescribeStacks(describeInput)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	if len(describeOutput.Stacks) != 1 {
		return nil, microerror.Mask(
			fmt.Errorf("unexpected number of stacks with name %q found, %d",
				stackName,
				len(describeOutput.Stacks)))
	}

	outputStackState := StackState{
		Name:    stackName,
		Outputs: describeOutput.Stacks[0].Outputs,
	}

	return outputStackState, nil
}
