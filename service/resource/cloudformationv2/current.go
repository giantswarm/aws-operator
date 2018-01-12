package cloudformationv2

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/service/keyv2"
)

func (r *Resource) GetCurrentState(ctx context.Context, obj interface{}) (interface{}, error) {
	customObject, err := keyv2.ToCustomObject(obj)
	if err != nil {
		return StackState{}, microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "debug", "looking for AWS stack")

	stackName := keyv2.MainGuestStackName(customObject)

	describeInput := &cloudformation.DescribeStacksInput{
		StackName: aws.String(stackName),
	}
	describeOutput, err := r.Clients.CloudFormation.DescribeStacks(describeInput)

	if IsStackNotFound(err) {
		r.logger.LogCtx(ctx, "debug", "did not find a stack in AWS API")
		return StackState{}, nil
	}
	if err != nil {
		return StackState{}, microerror.Mask(err)
	}

	if len(describeOutput.Stacks) > 1 {
		return StackState{}, microerror.Mask(notFoundError)
	}

	// current is called on cluster deletion, if the stack creation failed the
	// outputs can be unaccessible, this can lead to a stack that cannot be deleted.
	// it can also be called during creation, while the outputs are still not
	// accessible.
	status := describeOutput.Stacks[0].StackStatus
	errorStatuses := []string{
		"ROLLBACK_IN_PROGRESS",
		"ROLLBACK_COMPLETE",
		"CREATE_IN_PROGRESS",
	}
	for _, errorStatus := range errorStatuses {
		if *status == errorStatus {
			outputStackState := StackState{
				Name:           stackName,
				Workers:        "",
				ImageID:        "",
				ClusterVersion: "",
			}
			return outputStackState, nil
		}
	}

	outputs := describeOutput.Stacks[0].Outputs

	workers, err := getStackOutputValue(outputs, workersOutputKey)
	if err != nil {
		return StackState{}, microerror.Mask(err)
	}
	imageID, err := getStackOutputValue(outputs, imageIDOutputKey)
	if err != nil {
		return StackState{}, microerror.Mask(err)
	}
	clusterVersion, err := getStackOutputValue(outputs, clusterVersionOutputKey)
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
