package cloudformation

import (
	"context"
	"fmt"
	"strconv"

	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/controller/context/finalizerskeptcontext"
	"github.com/giantswarm/operatorkit/controller/context/resourcecanceledcontext"

	cloudformationservice "github.com/giantswarm/aws-operator/service/controller/v16/cloudformation"
	"github.com/giantswarm/aws-operator/service/controller/v16/controllercontext"
	"github.com/giantswarm/aws-operator/service/controller/v16/key"
)

func (r *Resource) GetCurrentState(ctx context.Context, obj interface{}) (interface{}, error) {
	customObject, err := key.ToCustomObject(obj)
	if err != nil {
		return StackState{}, microerror.Mask(err)
	}

	stackName := key.MainGuestStackName(customObject)

	sc, err := controllercontext.FromContext(ctx)
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
		stackOutputs, stackStatus, err = sc.CloudFormation.DescribeOutputsAndStatus(stackName)
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
		var hostedZoneNameServers string
		if r.route53Enabled {
			hostedZoneNameServers, err = sc.CloudFormation.GetOutputValue(stackOutputs, key.HostedZoneNameServers)
			// TODO introduced: aws-operator@v14; remove with: aws-operator@v13
			// This output was introduced in v14 so it isn't accessible from CF
			// stacks created by earlier versions. We need to handle that.
			//
			// Final version of the code:
			//
			//	if err != nil {
			//		return StackState{}, microerror.Mask(err)
			//	}
			//
			if cloudformationservice.IsOutputNotFound(err) {
				// Fall trough. Empty string is handled in host post stack creation.
			} else if err != nil {
				return StackState{}, microerror.Mask(err)
			}
			// TODO end
		}
		dockerVolumeResourceName, err := sc.CloudFormation.GetOutputValue(stackOutputs, key.DockerVolumeResourceNameKey)
		if cloudformationservice.IsOutputNotFound(err) {
			// Since we are transitioning between versions we will have situations in
			// which old clusters are updated to new versions and miss the docker
			// volume resource name in the CF stack outputs. We ignore this problem
			// for now and move on regardless. On the next resync period the output
			// value will be there, once the cluster got updated.
			//
			// TODO remove this condition as soon as all guest clusters in existence
			// obtain a docker volume resource.
			dockerVolumeResourceName = ""
		} else if err != nil {
			return StackState{}, microerror.Mask(err)
		}

		var workerDockerVolumeSizeGB int
		{
			v, err := sc.CloudFormation.GetOutputValue(stackOutputs, key.WorkerDockerVolumeSizeKey)
			if cloudformationservice.IsOutputNotFound(err) {
				// Since we are transitioning between versions we will have situations in
				// which old clusters are updated to new versions and miss the docker
				// volume resource name in the CF stack outputs. We ignore this problem
				// for now and move on regardless. On the next resync period the output
				// value will be there, once the cluster got updated.
				//
				// TODO remove this condition as soon as all guest clusters in existence
				// obtain a docker volume size. Tracked here: https://github.com/giantswarm/giantswarm/issues/4139.
				v = "100"
			} else if err != nil {
				return StackState{}, microerror.Mask(err)
			}

			sz, err := strconv.ParseUint(v, 10, 32)
			if err != nil {
				return StackState{}, microerror.Mask(err)
			}

			workerDockerVolumeSizeGB = int(sz)
		}

		masterImageID, err := sc.CloudFormation.GetOutputValue(stackOutputs, key.MasterImageIDKey)
		if err != nil {
			return StackState{}, microerror.Mask(err)
		}
		masterInstanceResourceName, err := sc.CloudFormation.GetOutputValue(stackOutputs, key.MasterInstanceResourceNameKey)
		if cloudformationservice.IsOutputNotFound(err) {
			// Since we are transitioning between versions we will have situations in
			// which old clusters are updated to new versions and miss the master
			// instance resource name in the CF stack outputs. We ignore this problem
			// for now and move on regardless. On the next resync period the output
			// value will be there, once the cluster got updated.
			//
			// TODO remove this condition as soon as all guest clusters in existence
			// obtain a master instance resource.
			masterInstanceResourceName = ""
		} else if err != nil {
			return StackState{}, microerror.Mask(err)
		}
		masterInstanceType, err := sc.CloudFormation.GetOutputValue(stackOutputs, key.MasterInstanceTypeKey)
		if err != nil {
			return StackState{}, microerror.Mask(err)
		}
		masterCloudConfigVersion, err := sc.CloudFormation.GetOutputValue(stackOutputs, key.MasterCloudConfigVersionKey)
		if err != nil {
			return StackState{}, microerror.Mask(err)
		}

		workerCount, err := sc.CloudFormation.GetOutputValue(stackOutputs, key.WorkerCountKey)
		if err != nil {
			return StackState{}, microerror.Mask(err)
		}
		workerImageID, err := sc.CloudFormation.GetOutputValue(stackOutputs, key.WorkerImageIDKey)
		if err != nil {
			return StackState{}, microerror.Mask(err)
		}
		workerInstanceType, err := sc.CloudFormation.GetOutputValue(stackOutputs, key.WorkerInstanceTypeKey)
		if err != nil {
			return StackState{}, microerror.Mask(err)
		}
		workerCloudConfigVersion, err := sc.CloudFormation.GetOutputValue(stackOutputs, key.WorkerCloudConfigVersionKey)
		if err != nil {
			return StackState{}, microerror.Mask(err)
		}

		versionBundleVersion, err := sc.CloudFormation.GetOutputValue(stackOutputs, key.VersionBundleVersionKey)
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

			HostedZoneNameServers: hostedZoneNameServers,

			DockerVolumeResourceName:   dockerVolumeResourceName,
			MasterImageID:              masterImageID,
			MasterInstanceResourceName: masterInstanceResourceName,
			MasterInstanceType:         masterInstanceType,
			MasterCloudConfigVersion:   masterCloudConfigVersion,

			WorkerCount:              workerCount,
			WorkerDockerVolumeSizeGB: workerDockerVolumeSizeGB,
			WorkerImageID:            workerImageID,
			WorkerInstanceType:       workerInstanceType,
			WorkerCloudConfigVersion: workerCloudConfigVersion,

			VersionBundleVersion: versionBundleVersion,
		}
	}

	return currentState, nil
}
