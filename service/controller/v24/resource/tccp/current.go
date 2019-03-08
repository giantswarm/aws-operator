package tccp

import (
	"context"
	"fmt"
	"strconv"

	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/controller/context/finalizerskeptcontext"
	"github.com/giantswarm/operatorkit/controller/context/resourcecanceledcontext"

	cloudformationservice "github.com/giantswarm/aws-operator/service/controller/v24/cloudformation"
	"github.com/giantswarm/aws-operator/service/controller/v24/controllercontext"
	"github.com/giantswarm/aws-operator/service/controller/v24/key"
)

func (r *Resource) GetCurrentState(ctx context.Context, obj interface{}) (interface{}, error) {
	customObject, err := key.ToCustomObject(obj)
	if err != nil {
		return StackState{}, microerror.Mask(err)
	}

	// The IPAM resource is executed before the CloudFormation resource in order
	// to allocate a free IP range for the tenant subnet. This CIDR is put into
	// the CR status. In case it is missing, the IPAM resource did not yet
	// allocate it and the CloudFormation resource cannot proceed. We cancel here
	// and wait for the CIDR to be available in the CR status.
	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "finding tenant subnet in CR status")

		if key.StatusNetworkCIDR(customObject) == "" {
			r.logger.LogCtx(ctx, "level", "debug", "message", "did not find tenant subnet in CR status")
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
			resourcecanceledcontext.SetCanceled(ctx)

			return StackState{}, microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "found tenant subnet in CR status")
	}

	stackName := key.MainGuestStackName(customObject)

	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return StackState{}, microerror.Mask(err)
	}

	// In order to compute the current state of the guest cluster's cloud
	// formation stack we have to describe the CF stacks and lookup the right
	// stack. We dispatch our custom StackState structure and enrich it with all
	// information necessary to reconcile the cloudformation resource.
	var stackOutputs []*cloudformation.Output
	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "finding the guest cluster main stack outputs in the AWS API")

		var stackStatus string
		stackOutputs, stackStatus, err = cc.CloudFormation.DescribeOutputsAndStatus(stackName)
		if cloudformationservice.IsStackNotFound(err) {
			r.logger.LogCtx(ctx, "level", "debug", "message", "did not find the guest cluster main stack outputs in the AWS API")
			r.logger.LogCtx(ctx, "level", "debug", "message", "the guest cluster main stack does not exist")
			return StackState{}, nil

		} else if cloudformationservice.IsOutputsNotAccessible(err) {
			r.logger.LogCtx(ctx, "level", "debug", "message", "did not find the guest cluster main stack outputs in the AWS API")
			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("the guest cluster main stack has status '%s'", stackStatus))
			if key.IsDeleted(customObject) {
				// Keep finalizers to as long as we don't
				// encounter IsStackNotFound.
				r.logger.LogCtx(ctx, "level", "debug", "message", "keeping finalizers")
				finalizerskeptcontext.SetKept(ctx)
			}
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
			resourcecanceledcontext.SetCanceled(ctx)

			return StackState{}, nil

		} else if err != nil {
			return StackState{}, microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "found the guest cluster main stack outputs in the AWS API")
	}

	var currentState StackState
	{
		dockerVolumeResourceName, err := cc.CloudFormation.GetOutputValue(stackOutputs, key.DockerVolumeResourceNameKey)
		if err != nil {
			return StackState{}, microerror.Mask(err)
		}

		masterImageID, err := cc.CloudFormation.GetOutputValue(stackOutputs, key.MasterImageIDKey)
		if err != nil {
			return StackState{}, microerror.Mask(err)
		}
		masterInstanceResourceName, err := cc.CloudFormation.GetOutputValue(stackOutputs, key.MasterInstanceResourceNameKey)
		if err != nil {
			return StackState{}, microerror.Mask(err)
		}
		masterInstanceType, err := cc.CloudFormation.GetOutputValue(stackOutputs, key.MasterInstanceTypeKey)
		if err != nil {
			return StackState{}, microerror.Mask(err)
		}
		masterCloudConfigVersion, err := cc.CloudFormation.GetOutputValue(stackOutputs, key.MasterCloudConfigVersionKey)
		if err != nil {
			return StackState{}, microerror.Mask(err)
		}

		workerCloudConfigVersion, err := cc.CloudFormation.GetOutputValue(stackOutputs, key.WorkerCloudConfigVersionKey)
		if err != nil {
			return StackState{}, microerror.Mask(err)
		}
		var workerDockerVolumeSizeGB int
		{
			v, err := cc.CloudFormation.GetOutputValue(stackOutputs, key.WorkerDockerVolumeSizeKey)
			if err != nil {
				return StackState{}, microerror.Mask(err)
			}

			workerDockerVolumeSizeGB, err = strconv.Atoi(v)
			if err != nil {
				return StackState{}, microerror.Mask(err)
			}
		}
		workerImageID, err := cc.CloudFormation.GetOutputValue(stackOutputs, key.WorkerImageIDKey)
		if err != nil {
			return StackState{}, microerror.Mask(err)
		}
		workerInstanceType, err := cc.CloudFormation.GetOutputValue(stackOutputs, key.WorkerInstanceTypeKey)
		if err != nil {
			return StackState{}, microerror.Mask(err)
		}

		versionBundleVersion, err := cc.CloudFormation.GetOutputValue(stackOutputs, key.VersionBundleVersionKey)
		if err != nil {
			return StackState{}, microerror.Mask(err)
		}

		currentState = StackState{
			Name: stackName,

			DockerVolumeResourceName: dockerVolumeResourceName,

			MasterImageID:              masterImageID,
			MasterInstanceResourceName: masterInstanceResourceName,
			MasterInstanceType:         masterInstanceType,
			MasterCloudConfigVersion:   masterCloudConfigVersion,

			WorkerCloudConfigVersion: workerCloudConfigVersion,
			WorkerDockerVolumeSizeGB: workerDockerVolumeSizeGB,
			WorkerImageID:            workerImageID,
			WorkerInstanceType:       workerInstanceType,

			VersionBundleVersion: versionBundleVersion,
		}
	}

	return currentState, nil
}
