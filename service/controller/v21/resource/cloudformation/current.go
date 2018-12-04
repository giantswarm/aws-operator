package cloudformation

import (
	"context"
	"fmt"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/controller/context/finalizerskeptcontext"
	"github.com/giantswarm/operatorkit/controller/context/resourcecanceledcontext"

	cloudformationservice "github.com/giantswarm/aws-operator/service/controller/v21/cloudformation"
	"github.com/giantswarm/aws-operator/service/controller/v21/controllercontext"
	"github.com/giantswarm/aws-operator/service/controller/v21/key"
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

		if key.ClusterNetworkCIDR(customObject) == "" {
			r.logger.LogCtx(ctx, "level", "debug", "message", "did not find tenant subnet in CR status")
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
			resourcecanceledcontext.SetCanceled(ctx)

			return StackState{}, microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "found tenant subnet in CR status")
	}

	stackName := key.MainGuestStackName(customObject)

	ctlCtx, err := controllercontext.FromContext(ctx)
	if err != nil {
		return StackState{}, microerror.Mask(err)
	}

	if key.IsDeleted(customObject) {
		stackNames := []string{
			key.MainGuestStackName(customObject),
			key.MainHostPreStackName(customObject),
			key.MainHostPostStackName(customObject),
		}

		for _, stackName := range stackNames {
			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("finding stack %#q in the AWS API", stackName))

			in := &cloudformation.DescribeStacksInput{
				StackName: aws.String(stackName),
			}

			_, err := ctlCtx.AWSClient.CloudFormation.DescribeStacks(in)
			if cloudformationservice.IsStackNotFound(err) {
				// This handling is far from perfect. We use different
				// packages here. This is all going to be addressed in
				// scope of
				// https://github.com/giantswarm/giantswarm/issues/3783.

				r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("did not find stack %#q in the AWS API", stackName))
			} else if err != nil {
				return nil, microerror.Mask(err)
			} else {
				r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("found stack %#q in the AWS API", stackName))
				r.logger.LogCtx(ctx, "level", "debug", "message", "keeping finalizer")
				finalizerskeptcontext.SetKept(ctx)
			}
		}

		// When a tenant cluster is deleted it might be not completely created yet
		// in the first place. There can be issues with unaccessible stack output
		// values in such cases, causing the deletion process to get into a
		// deadlock. To remedy such cases we simply return the stack state
		// containing the stack name, without trying to access any stack output
		// values.
		currentState := StackState{
			Name: key.MainGuestStackName(customObject),
		}

		return currentState, nil
	}

	// In order to compute the current state of the guest cluster's cloud
	// formation stack we have to describe the CF stacks and lookup the right
	// stack. We dispatch our custom StackState structure and enrich it with all
	// information necessary to reconcile the cloudformation resource.
	var stackOutputs []*cloudformation.Output
	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "finding the guest cluster main stack outputs in the AWS API")

		var stackStatus string
		stackOutputs, stackStatus, err = ctlCtx.CloudFormation.DescribeOutputsAndStatus(stackName)
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
			hostedZoneNameServers, err = ctlCtx.CloudFormation.GetOutputValue(stackOutputs, key.HostedZoneNameServers)
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
		dockerVolumeResourceName, err := ctlCtx.CloudFormation.GetOutputValue(stackOutputs, key.DockerVolumeResourceNameKey)
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
			v, err := ctlCtx.CloudFormation.GetOutputValue(stackOutputs, key.WorkerDockerVolumeSizeKey)
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

		masterImageID, err := ctlCtx.CloudFormation.GetOutputValue(stackOutputs, key.MasterImageIDKey)
		if err != nil {
			return StackState{}, microerror.Mask(err)
		}
		masterInstanceResourceName, err := ctlCtx.CloudFormation.GetOutputValue(stackOutputs, key.MasterInstanceResourceNameKey)
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
		masterInstanceType, err := ctlCtx.CloudFormation.GetOutputValue(stackOutputs, key.MasterInstanceTypeKey)
		if err != nil {
			return StackState{}, microerror.Mask(err)
		}
		masterCloudConfigVersion, err := ctlCtx.CloudFormation.GetOutputValue(stackOutputs, key.MasterCloudConfigVersionKey)
		if err != nil {
			return StackState{}, microerror.Mask(err)
		}

		workerCount, err := ctlCtx.CloudFormation.GetOutputValue(stackOutputs, key.WorkerCountKey)
		if err != nil {
			return StackState{}, microerror.Mask(err)
		}
		workerImageID, err := ctlCtx.CloudFormation.GetOutputValue(stackOutputs, key.WorkerImageIDKey)
		if err != nil {
			return StackState{}, microerror.Mask(err)
		}
		workerInstanceType, err := ctlCtx.CloudFormation.GetOutputValue(stackOutputs, key.WorkerInstanceTypeKey)
		if err != nil {
			return StackState{}, microerror.Mask(err)
		}
		workerCloudConfigVersion, err := ctlCtx.CloudFormation.GetOutputValue(stackOutputs, key.WorkerCloudConfigVersionKey)
		if err != nil {
			return StackState{}, microerror.Mask(err)
		}

		versionBundleVersion, err := ctlCtx.CloudFormation.GetOutputValue(stackOutputs, key.VersionBundleVersionKey)
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
