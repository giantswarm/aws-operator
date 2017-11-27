package cloudformation

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	awscloudformation "github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/service/key"
)

func (r *Resource) GetCurrentState(ctx context.Context, obj interface{}) (interface{}, error) {
	customObject, err := key.ToCustomObject(obj)
	if err != nil {
		return StackState{}, microerror.Mask(err)
	}
	// no-op if we are not using cloudformation
	if !key.UseCloudFormation(customObject) {
		r.logger.LogCtx(ctx, "debug", "not processing cloudformation stack")
		return StackState{}, nil
	}

	r.logger.LogCtx(ctx, "debug", "looking for AWS stack")

	stackName := key.MainStackName(customObject)

	describeInput := &awscloudformation.DescribeStacksInput{
		StackName: aws.String(stackName),
	}
	describeOutput, err := r.awsClients.CloudFormation.DescribeStacks(describeInput)

	if IsStackNotFound(err) {

		r.logger.LogCtx(ctx, "debug", "did not find a stack in AWS API")
		return StackState{}, nil

	} else if err != nil {

		return StackState{}, microerror.Mask(err)

	}

	if len(describeOutput.Stacks) > 1 {
		return StackState{}, microerror.Mask(notFoundError)
	}

	outputs := describeOutput.Stacks[0].Outputs

	workers, err := getStackOutputValue(outputs, workersParameterKey)
	if err != nil {
		return StackState{}, microerror.Mask(err)
	}
	imageID, err := getStackOutputValue(outputs, imageIDParameterKey)
	if err != nil {
		return StackState{}, microerror.Mask(err)
	}
	clusterVersion, err := getStackOutputValue(outputs, clusterVersionParameterKey)
	if err != nil {
		return StackState{}, microerror.Mask(err)
	}

	outputStackState := StackState{
		Name:           stackName,
		Workers:        workers,
		ImageID:        imageID,
		ClusterVersion: clusterVersion,
	}

	return outputStackState, nil
}
