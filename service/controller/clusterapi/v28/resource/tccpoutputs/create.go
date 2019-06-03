package tccpoutputs

import (
	"context"
	"fmt"

	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v28/cloudformation"
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v28/controllercontext"
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v28/key"
)

const (
	DockerVolumeResourceNameKey   = "DockerVolumeResourceName"
	HostedZoneNameServersKey      = "HostedZoneNameServers"
	MasterCloudConfigVersionKey   = "MasterCloudConfigVersion"
	MasterImageIDKey              = "MasterImageID"
	MasterInstanceResourceNameKey = "MasterInstanceResourceName"
	MasterInstanceTypeKey         = "MasterInstanceType"
	VersionBundleVersionKey       = "VersionBundleVersion"
	VPCIDKey                      = "VPCID"
	VPCPeeringConnectionIDKey     = "VPCPeeringConnectionID"
	WorkerASGNameKey              = "WorkerASGName"
	WorkerCloudConfigVersionKey   = "WorkerCloudConfigVersion"
	WorkerDockerVolumeSizeKey     = "WorkerDockerVolumeSizeGB"
	WorkerImageIDKey              = "WorkerImageID"
	WorkerInstanceTypeKey         = "WorkerInstanceType"
)

func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	cr, err := key.ToCluster(obj)
	if err != nil {
		return microerror.Mask(err)
	}
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	var cloudFormation *cloudformation.CloudFormation
	{
		c := cloudformation.Config{
			Client: cc.Client.TenantCluster.AWS.CloudFormation,
		}

		cloudFormation, err = cloudformation.New(c)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	var outputs []cloudformation.Output
	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "finding the tenant cluster cloud formation stack outputs")

		o, s, err := cloudFormation.DescribeOutputsAndStatus(key.StackNameTCCP(cr))
		if cloudformation.IsStackNotFound(err) {
			r.logger.LogCtx(ctx, "level", "debug", "message", "did not find the tenant cluster cloud formation stack outputs")
			r.logger.LogCtx(ctx, "level", "debug", "message", "the tenant cluster cloud formation stack does not exist")
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
			return nil

		} else if cloudformation.IsOutputsNotAccessible(err) {
			r.logger.LogCtx(ctx, "level", "debug", "message", "did not find the tenant cluster cloud formation stack outputs")
			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("the tenant cluster main cloud formation stack output values are not accessible due to stack status %#q", s))
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
			cc.Status.TenantCluster.TCCP.IsTransitioning = true
			return nil

		} else if err != nil {
			return microerror.Mask(err)
		}

		outputs = o

		r.logger.LogCtx(ctx, "level", "debug", "message", "found the tenant cluster cloud formation stack outputs")
	}

	{
		v, err := cloudFormation.GetOutputValue(outputs, DockerVolumeResourceNameKey)
		if err != nil {
			return microerror.Mask(err)
		}
		cc.Status.TenantCluster.MasterInstance.DockerVolumeResourceName = v
	}

	if r.route53Enabled {
		v, err := cloudFormation.GetOutputValue(outputs, HostedZoneNameServersKey)
		if err != nil {
			return microerror.Mask(err)
		}
		cc.Status.TenantCluster.HostedZoneNameServers = v
	}

	{
		v, err := cloudFormation.GetOutputValue(outputs, MasterImageIDKey)
		if err != nil {
			return microerror.Mask(err)
		}
		cc.Status.TenantCluster.MasterInstance.Image = v
	}

	{
		v, err := cloudFormation.GetOutputValue(outputs, MasterInstanceResourceNameKey)
		if err != nil {
			return microerror.Mask(err)
		}
		cc.Status.TenantCluster.MasterInstance.ResourceName = v
	}

	{
		v, err := cloudFormation.GetOutputValue(outputs, MasterInstanceTypeKey)
		if err != nil {
			return microerror.Mask(err)
		}
		cc.Status.TenantCluster.MasterInstance.Type = v
	}

	{
		v, err := cloudFormation.GetOutputValue(outputs, MasterCloudConfigVersionKey)
		if err != nil {
			return microerror.Mask(err)
		}
		cc.Status.TenantCluster.MasterInstance.CloudConfigVersion = v
	}

	{
		v, err := cloudFormation.GetOutputValue(outputs, VersionBundleVersionKey)
		if err != nil {
			return microerror.Mask(err)
		}
		cc.Status.TenantCluster.VersionBundleVersion = v
	}

	{
		v, err := cloudFormation.GetOutputValue(outputs, VPCIDKey)
		if err != nil {
			return microerror.Mask(err)
		}
		cc.Status.TenantCluster.TCCP.VPC.ID = v
	}

	{
		v, err := cloudFormation.GetOutputValue(outputs, VPCPeeringConnectionIDKey)
		if err != nil {
			return microerror.Mask(err)
		}
		cc.Status.TenantCluster.TCCP.VPC.PeeringConnectionID = v
	}

	{
		v, err := cloudFormation.GetOutputValue(outputs, WorkerCloudConfigVersionKey)
		if err != nil {
			return microerror.Mask(err)
		}
		cc.Status.TenantCluster.WorkerInstance.CloudConfigVersion = v
	}

	{
		v, err := cloudFormation.GetOutputValue(outputs, WorkerDockerVolumeSizeKey)
		if err != nil {
			return microerror.Mask(err)
		}
		cc.Status.TenantCluster.WorkerInstance.DockerVolumeSizeGB = v
	}

	{
		v, err := cloudFormation.GetOutputValue(outputs, WorkerImageIDKey)
		if err != nil {
			return microerror.Mask(err)
		}
		cc.Status.TenantCluster.WorkerInstance.Image = v
	}

	{
		v, err := cloudFormation.GetOutputValue(outputs, WorkerInstanceTypeKey)
		if err != nil {
			return microerror.Mask(err)
		}
		cc.Status.TenantCluster.WorkerInstance.Type = v
	}

	return nil
}
