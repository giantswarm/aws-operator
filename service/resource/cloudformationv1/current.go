package cloudformationv1

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

	r.logger.LogCtx(ctx, "debug", "looking for AWS stack")

	stackName := key.MainStackName(customObject)

	describeInput := &awscloudformation.DescribeStacksInput{
		StackName: aws.String(stackName),
	}
	describeOutput, err := r.awsClients.CloudFormation.DescribeStacks(describeInput)

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

	// current is called on cluster deletion, if the stakc creation failed the
	// outputs can be unaccessible, this can lead to a stack that cannot be deleted.
	status := describeOutput.Stacks[0].StackStatus
	errorStatuses := []string{
		"ROLLBACK_IN_PROGRESS",
		"ROLLBACK_COMPLETE",
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
