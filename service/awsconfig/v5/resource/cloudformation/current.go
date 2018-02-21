package cloudformation

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/framework/context/resourcecanceledcontext"

	"github.com/giantswarm/aws-operator/service/awsconfig/v5/key"
)

func (r *Resource) GetCurrentState(ctx context.Context, obj interface{}) (interface{}, error) {
	customObject, err := key.ToCustomObject(obj)
	if err != nil {
		return StackState{}, microerror.Mask(err)
	}

	// In order to compute the current state of the guest cluster's main stack we
	// have to describe the CF stacks and lookup the right stack. We dispatch our
	// custom StackState structure and enrich it with all information necessary to
	// reconcile the cloudformation resource.
	stackName := key.MainGuestStackName(customObject)
	describeOutput := &cloudformation.DescribeStacksOutput{}
	{
		r.logger.LogCtx(ctx, "debug", "looking for main stack")

		describeInput := &cloudformation.DescribeStacksInput{
			StackName: aws.String(stackName),
		}
		describeOutput, err = r.clients.CloudFormation.DescribeStacks(describeInput)
		if IsStackNotFound(err) {
			r.logger.LogCtx(ctx, "debug", "did not find main stack")
			return StackState{}, nil
		} else if err != nil {
			return StackState{}, microerror.Mask(err)
		}
		if len(describeOutput.Stacks) > 1 {
			return StackState{}, microerror.Mask(notFoundError)
		}

		r.logger.LogCtx(ctx, "debug", "found main stack")
	}

	// GetCurrentState is called on cluster deletion, if the stack creation failed
	// the outputs can be unaccessible, this can lead to a stack that cannot be
	// deleted. it can also be called during creation, while the outputs are still
	// not accessible.
	status := *describeOutput.Stacks[0].StackStatus
	{
		errorStatuses := []string{
			"ROLLBACK_IN_PROGRESS",
			"ROLLBACK_COMPLETE",
			"CREATE_IN_PROGRESS",
		}

		for _, errorStatus := range errorStatuses {
			if status == errorStatus {
				outputStackState := StackState{
					Name: stackName,
				}

				return outputStackState, nil
			}
		}
	}

	// In case the current guest cluster is already being updated, we cancel the
	// reconciliation until the current update is done in order to reduce
	// unnecessary friction.
	if status == cloudformation.ResourceStatusUpdateInProgress {
		r.logger.LogCtx(ctx, "debug", fmt.Sprintf("main stack is in state '%s'", cloudformation.ResourceStatusUpdateInProgress))
		resourcecanceledcontext.SetCanceled(ctx)
		r.logger.LogCtx(ctx, "debug", "canceling resource for custom object")

		return StackState{}, nil
	}

	// Here we finally construct the current state.
	currentState := StackState{}
	{
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
		versionBundleVersion, err := getStackOutputValue(outputs, versionBundleVersionOutputKey)
		if IsNotFound(err) {
			// Since we are transitioning between versions we will have situations in
			// which old clusters are updated to new versions and miss the version
			// bundle version in the CF stack outputs. We ignore this problem for now
			// and move on regardless. The reconciliation will detect the guest cluster
			// needs to be updated and once this is done, we should be fine again.
			//
			// TODO remove this condition as soon as all guest clusters in existence
			// obtain a version bundle version.
			versionBundleVersion = ""
		} else if err != nil {
			return StackState{}, microerror.Mask(err)
		}

		currentState = StackState{
			Name: stackName,

			MasterImageID:            masterImageID,
			MasterInstanceType:       masterInstanceType,
			MasterCloudConfigVersion: masterCloudConfigVersion,

			WorkerCount:              workers,
			WorkerImageID:            workerImageID,
			WorkerInstanceType:       workerInstanceType,
			WorkerCloudConfigVersion: workerCloudConfigVersion,

			VersionBundleVersion: versionBundleVersion,
		}
	}

	return currentState, nil
}
