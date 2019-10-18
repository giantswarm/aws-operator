package tccpoutputs

import (
	"context"
	"fmt"
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v31/cloudformation"
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v31/controllercontext"
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v31/key"
)

const (
	DockerVolumeResourceNameKey   = "DockerVolumeResourceName"
	HostedZoneNameServersKey      = "HostedZoneNameServers"
	IngressInsecureTargetGroupIDs = "IngressInsecureTargetGroupsID"
	IngressSecureTargetGroupIDs   = "IngressSecureTargetGroupsID"
	MasterImageIDKey              = "MasterImageID"
	MasterInstanceResourceNameKey = "MasterInstanceResourceName"
	MasterInstanceTypeKey         = "MasterInstanceType"
	OperatorVersion               = "OperatorVersion"
	VPCIDKey                      = "VPCID"
	VPCPeeringConnectionIDKey     = "VPCPeeringConnectionID"
)

func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	cr, err := r.toClusterFunc(obj)
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
		r.logger.LogCtx(ctx, "level", "debug", "message", "finding the tenant cluster's control plane cloud formation stack outputs")

		o, s, err := cloudFormation.DescribeOutputsAndStatus(key.StackNameTCCP(&cr))
		if cloudformation.IsStackNotFound(err) {
			r.logger.LogCtx(ctx, "level", "debug", "message", "did not find the tenant cluster's control plane cloud formation stack outputs")
			r.logger.LogCtx(ctx, "level", "debug", "message", "the tenant cluster's control plane cloud formation stack does not exist")
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
			return nil

		} else if cloudformation.IsOutputsNotAccessible(err) {
			r.logger.LogCtx(ctx, "level", "debug", "message", "did not find the tenant cluster's control plane cloud formation stack outputs")
			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("the tenant cluster's control plane cloud formation stack output values are not accessible due to stack status %#q", s))
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
			cc.Status.TenantCluster.TCCP.IsTransitioning = true
			return nil

		} else if err != nil {
			return microerror.Mask(err)
		}

		outputs = o

		r.logger.LogCtx(ctx, "level", "debug", "message", "found the tenant cluster's control plane cloud formation stack outputs")
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
		ingressInsecureTargetGroup, err := cloudFormation.GetOutputValue(outputs, IngressInsecureTargetGroupIDs)
		if err != nil {
			return microerror.Mask(err)
		}
		ingressSecureTargetGroup, err := cloudFormation.GetOutputValue(outputs, IngressSecureTargetGroupIDs)
		if err != nil {
			return microerror.Mask(err)
		}
		cc.Status.TenantCluster.TCCP.IngressTargetGroupIDs = []string{ingressInsecureTargetGroup, ingressSecureTargetGroup}
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
		v, err := cloudFormation.GetOutputValue(outputs, OperatorVersion)
		if err != nil {
			return microerror.Mask(err)
		}
		cc.Status.TenantCluster.OperatorVersion = v
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

	return nil
}
