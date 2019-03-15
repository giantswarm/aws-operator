package tccp

import (
	"context"
	"fmt"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/controller/context/finalizerskeptcontext"
	"github.com/giantswarm/operatorkit/controller/context/resourcecanceledcontext"

	cf "github.com/giantswarm/aws-operator/service/controller/v25/cloudformation"
	"github.com/giantswarm/aws-operator/service/controller/v25/controllercontext"
	"github.com/giantswarm/aws-operator/service/controller/v25/key"
)

func (r *Resource) GetCurrentState(ctx context.Context, obj interface{}) (interface{}, error) {
	cr, err := key.ToCustomObject(obj)
	if err != nil {
		return StackState{}, microerror.Mask(err)
	}
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return StackState{}, microerror.Mask(err)
	}

	var cloudFormation *cf.CloudFormation
	{
		c := cf.Config{
			Client: cc.Client.TenantCluster.AWS.CloudFormation,
		}

		cloudFormation, err = cf.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	// The IPAM resource is executed before the CloudFormation resource in order
	// to allocate a free IP range for the tenant subnet. This CIDR is put into
	// the CR status. In case it is missing, the IPAM resource did not yet
	// allocate it and the CloudFormation resource cannot proceed. We cancel here
	// and wait for the CIDR to be available in the CR status.
	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "finding tenant subnet in CR status")

		if key.StatusNetworkCIDR(cr) == "" {
			r.logger.LogCtx(ctx, "level", "debug", "message", "did not find tenant subnet in CR status")
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
			resourcecanceledcontext.SetCanceled(ctx)

			return StackState{}, microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "found tenant subnet in CR status")
	}

	// In order to compute the current state of the tenant cluster's cloud
	// formation stack we have to describe the CF stacks and lookup the right
	// stack. We dispatch our custom StackState structure and enrich it with all
	// information necessary to reconcile the cloudformation resource.
	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "finding the tenant cluster main stack outputs in the AWS API")

		var stackStatus string
		_, stackStatus, err := cloudFormation.DescribeOutputsAndStatus(key.MainGuestStackName(cr))
		if cf.IsStackNotFound(err) {
			r.logger.LogCtx(ctx, "level", "debug", "message", "did not find the tenant cluster main stack outputs in the AWS API")
			r.logger.LogCtx(ctx, "level", "debug", "message", "the tenant cluster main stack does not exist")
			return StackState{}, nil

		} else if cf.IsOutputsNotAccessible(err) {
			r.logger.LogCtx(ctx, "level", "debug", "message", "did not find the tenant cluster main stack outputs in the AWS API")
			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("the tenant cluster main stack has status '%s'", stackStatus))

			if key.IsDeleted(cr) {
				// Keep finalizers to as long as we don't
				// encounter IsStackNotFound.
				r.logger.LogCtx(ctx, "level", "debug", "message", "keeping finalizers")
				finalizerskeptcontext.SetKept(ctx)
			} else {
				r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
				resourcecanceledcontext.SetCanceled(ctx)
			}

			return StackState{}, nil

		} else if err != nil {
			return StackState{}, microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "found the tenant cluster main stack outputs in the AWS API")
	}

	var currentState StackState
	{
		currentState = StackState{
			Name: key.MainGuestStackName(cr),

			DockerVolumeResourceName:   cc.Status.TenantCluster.MasterInstance.DockerVolumeResourceName,
			MasterImageID:              cc.Status.TenantCluster.MasterInstance.Image,
			MasterInstanceResourceName: cc.Status.TenantCluster.MasterInstance.ResourceName,
			MasterInstanceType:         cc.Status.TenantCluster.MasterInstance.Type,
			MasterCloudConfigVersion:   cc.Status.TenantCluster.MasterInstance.CloudConfigVersion,

			WorkerCloudConfigVersion: cc.Status.TenantCluster.WorkerInstance.CloudConfigVersion,
			WorkerDockerVolumeSizeGB: cc.Status.TenantCluster.WorkerInstance.DockerVolumeSizeGB,
			WorkerImageID:            cc.Status.TenantCluster.WorkerInstance.Image,
			WorkerInstanceType:       cc.Status.TenantCluster.WorkerInstance.Type,

			VersionBundleVersion: cc.Status.TenantCluster.VersionBundleVersion,
		}
	}

	return currentState, nil
}
