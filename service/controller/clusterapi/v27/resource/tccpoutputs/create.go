package tccpoutputs

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v27/cloudformation"
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v27/controllercontext"
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v27/key"
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
		if cloudformation.IsOutputNotFound(err) {
			// TODO this exception is necessary for clusters upgrading from v24 to
			// v25. The code can be cleaned up in v27 and the controller context value
			// assignment can be managed like the other examples below.
			//
			//     https://github.com/giantswarm/giantswarm/issues/5570
			//
			v, err := searchVPCID(cc.Client.TenantCluster.AWS.EC2, key.ClusterID(cr))
			if err != nil {
				return microerror.Mask(err)
			}
			cc.Status.TenantCluster.TCCP.VPC.ID = v
		} else if err != nil {
			return microerror.Mask(err)
		} else {
			cc.Status.TenantCluster.TCCP.VPC.ID = v
		}
	}

	{
		v, err := cloudFormation.GetOutputValue(outputs, VPCPeeringConnectionIDKey)
		if cloudformation.IsOutputNotFound(err) {
			// TODO this exception is necessary for clusters upgrading from v23 to
			// v24. The code can be cleaned up in v25 and the controller context value
			// assignment can be managed like the other examples below.
			//
			//     https://github.com/giantswarm/giantswarm/issues/5496
			//
			v, err := searchPeeringConnectionID(cc.Client.TenantCluster.AWS.EC2, key.ClusterID(cr))
			if err != nil {
				return microerror.Mask(err)
			}
			cc.Status.TenantCluster.TCCP.VPC.PeeringConnectionID = v
		} else if err != nil {
			return microerror.Mask(err)
		} else {
			cc.Status.TenantCluster.TCCP.VPC.PeeringConnectionID = v
		}
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

func searchPeeringConnectionID(client EC2, clusterID string) (string, error) {
	var peeringID string
	{
		i := &ec2.DescribeVpcPeeringConnectionsInput{
			Filters: []*ec2.Filter{
				{
					Name: aws.String("status-code"),
					Values: []*string{
						aws.String("active"),
					},
				},
				{
					Name: aws.String("tag:Name"),
					Values: []*string{
						aws.String(clusterID),
					},
				},
			},
		}

		o, err := client.DescribeVpcPeeringConnections(i)
		if err != nil {
			return "", microerror.Mask(err)
		}
		if len(o.VpcPeeringConnections) != 1 {
			return "", microerror.Maskf(executionFailedError, "expected one vpc peering connection, got %d", len(o.VpcPeeringConnections))
		}

		peeringID = *o.VpcPeeringConnections[0].VpcPeeringConnectionId
	}

	return peeringID, nil
}

func searchVPCID(client EC2, clusterID string) (string, error) {
	var vpcID string
	{
		i := &ec2.DescribeVpcsInput{
			Filters: []*ec2.Filter{
				{
					Name: aws.String("tag:Name"),
					Values: []*string{
						aws.String(clusterID),
					},
				},
			},
		}

		o, err := client.DescribeVpcs(i)
		if err != nil {
			return "", microerror.Mask(err)
		}
		if len(o.Vpcs) != 1 {
			return "", microerror.Maskf(executionFailedError, "expected one vpc, got %d", len(o.Vpcs))
		}

		vpcID = *o.Vpcs[0].VpcId
	}

	return vpcID, nil
}
