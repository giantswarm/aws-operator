package cloudformation

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/service/awsconfig/v5/key"
)

func (r *Resource) GetCurrentState(ctx context.Context, obj interface{}) (interface{}, error) {
	customObject, err := key.ToCustomObject(obj)
	if err != nil {
		return StackState{}, microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "debug", "looking for AWS stack")

	stackName := key.MainGuestStackName(customObject)

	describeInput := &cloudformation.DescribeStacksInput{
		StackName: aws.String(stackName),
	}
	describeOutput, err := r.clients.CloudFormation.DescribeStacks(describeInput)
	if IsStackNotFound(err) {
		r.logger.LogCtx(ctx, "debug", "did not find a stack in AWS API")
		return StackState{}, nil
	} else if err != nil {
		return StackState{}, microerror.Mask(err)
	}

	if len(describeOutput.Stacks) > 1 {
		return StackState{}, microerror.Mask(notFoundError)
	}

	// GetCurrentState is called on cluster deletion, if the stack creation failed
	// the outputs can be unaccessible, this can lead to a stack that cannot be
	// deleted. it can also be called during creation, while the outputs are still
	// not accessible.
	status := describeOutput.Stacks[0].StackStatus
	errorStatuses := []string{
		"ROLLBACK_IN_PROGRESS",
		"ROLLBACK_COMPLETE",
		"CREATE_IN_PROGRESS",
	}
	for _, errorStatus := range errorStatuses {
		if *status == errorStatus {
			outputStackState := StackState{
				Name: stackName,
			}
			return outputStackState, nil
		}
	}

	outputs := describeOutput.Stacks[0].Outputs

	masterImageID, err := getStackOutputValue(outputs, masterImageIDOutputKey)
	if err != nil {
		return StackState{}, microerror.Mask(err)
	}
	masterInstanceType, err := getStackOutputValue(outputs, masterInstanceTypeOutputKey)
	if err != nil {
		return StackState{}, microerror.Mask(err)
	}
	masterCloudConfigVersion, err := getStackOutputValue(outputs, masterCloudConfigVersionOutputKey)
	if err != nil {
		return StackState{}, microerror.Mask(err)
	}
	workers, err := getStackOutputValue(outputs, workersOutputKey)
	if err != nil {
		return StackState{}, microerror.Mask(err)
	}
	workerImageID, err := getStackOutputValue(outputs, workerImageIDOutputKey)
	if err != nil {
		return StackState{}, microerror.Mask(err)
	}
	workerInstanceType, err := getStackOutputValue(outputs, workerInstanceTypeOutputKey)
	if err != nil {
		return StackState{}, microerror.Mask(err)
	}
	workerCloudConfigVersion, err := getStackOutputValue(outputs, workerCloudConfigVersionOutputKey)
	if err != nil {
		return StackState{}, microerror.Mask(err)
	}

	outputStackState := StackState{
		Name:                     stackName,
		MasterImageID:            masterImageID,
		MasterInstanceType:       masterInstanceType,
		MasterCloudConfigVersion: masterCloudConfigVersion,
		WorkerCount:              workers,
		WorkerImageID:            workerImageID,
		WorkerInstanceType:       workerInstanceType,
		WorkerCloudConfigVersion: workerCloudConfigVersion,
	}

	return outputStackState, nil
}
