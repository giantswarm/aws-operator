package cloudformation

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/framework/context/resourcecanceledcontext"

	cloudformationservice "github.com/giantswarm/aws-operator/service/awsconfig/v8/cloudformation"
	"github.com/giantswarm/aws-operator/service/awsconfig/v8/key"
)

func (r *Resource) GetCurrentState(ctx context.Context, obj interface{}) (interface{}, error) {
	customObject, err := key.ToCustomObject(obj)
	if err != nil {
		return StackState{}, microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", "looking for the guest cluster main stack in the AWS API")

	stackName := key.MainGuestStackName(customObject)

	// In order to compute the current state of the guest cluster's cloud
	// formation stack we have to describe the CF stacks and lookup the right
	// stack. We dispatch our custom StackState structure and enrich it with all
	// information necessary to reconcile the cloudformation resource.
	var stackOutputs []*cloudformation.Output
	var stackStatus string
	{
		stackOutputs, stackStatus, err = r.service.DescribeOutputsAndStatus(stackName)
		if cloudformationservice.IsStackNotFound(err) {
			r.logger.LogCtx(ctx, "level", "debug", "message", "did not find the guest cluster main stack in the AWS API")
			return StackState{}, nil

		} else if cloudformationservice.IsOutputsNotAccessible(err) {
			r.logger.LogCtx(ctx, "level", "debug", "message", "the guest cluster main stack output values are not accessible due to stack state transition")
			return StackState{Name: stackName}, nil

		} else if err != nil {
			return StackState{}, microerror.Mask(err)
		}
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", "found the guest cluster main stack in the AWS API")

	// In case the current guest cluster is already being updated, we cancel the
	// reconciliation until the current update is done in order to reduce
	// unnecessary friction.
	if stackStatus == cloudformation.ResourceStatusUpdateInProgress {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("guest cluster main stack is in state '%s'", cloudformation.ResourceStatusUpdateInProgress))
		resourcecanceledcontext.SetCanceled(ctx)
		r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource reconciliation for custom object")

		return StackState{}, nil
	}

	var currentState StackState
	{
		masterImageID, err := r.service.GetOutputValue(stackOutputs, key.MasterImageIDKey)
		if err != nil {
			return StackState{}, microerror.Mask(err)
		}
		masterInstanceType, err := r.service.GetOutputValue(stackOutputs, key.MasterInstanceTypeKey)
		if err != nil {
			return StackState{}, microerror.Mask(err)
		}
		masterCloudConfigVersion, err := r.service.GetOutputValue(stackOutputs, key.MasterCloudConfigVersionKey)
		if err != nil {
			return StackState{}, microerror.Mask(err)
		}
		workerCount, err := r.service.GetOutputValue(stackOutputs, key.WorkerCountKey)
		if err != nil {
			return StackState{}, microerror.Mask(err)
		}
		workerImageID, err := r.service.GetOutputValue(stackOutputs, key.WorkerImageIDKey)
		if err != nil {
			return StackState{}, microerror.Mask(err)
		}
		workerInstanceType, err := r.service.GetOutputValue(stackOutputs, key.WorkerInstanceTypeKey)
		if err != nil {
			return StackState{}, microerror.Mask(err)
		}
		workerCloudConfigVersion, err := r.service.GetOutputValue(stackOutputs, key.WorkerCloudConfigVersionKey)
		if err != nil {
			return StackState{}, microerror.Mask(err)
		}
		versionBundleVersion, err := r.service.GetOutputValue(stackOutputs, key.VersionBundleVersionKey)
		if cloudformationservice.IsOutputNotFound(err) {
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

			WorkerCount:              workerCount,
			WorkerImageID:            workerImageID,
			WorkerInstanceType:       workerInstanceType,
			WorkerCloudConfigVersion: workerCloudConfigVersion,

			VersionBundleVersion: versionBundleVersion,
		}
	}

	return currentState, nil
}
