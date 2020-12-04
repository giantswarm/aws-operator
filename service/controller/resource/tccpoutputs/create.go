package tccpoutputs

import (
	"context"

	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/service/controller/controllercontext"
	"github.com/giantswarm/aws-operator/service/controller/key"
	"github.com/giantswarm/aws-operator/service/internal/cloudformation"
)

const (
	APIServerPublicLoadBalancerKey = "APIServerPublicLoadBalancer"
	HostedZoneID                   = "HostedZoneID"
	HostedZoneNameServersKey       = "HostedZoneNameServers"
	InternalHostedZoneID           = "InternalHostedZoneID"
	OperatorVersion                = "OperatorVersion"
	VPCIDKey                       = "VPCID"
	VPCPeeringConnectionIDKey      = "VPCPeeringConnectionID"
)

func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	cr, err := r.toClusterFunc(ctx, obj)
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
		r.logger.Debugf(ctx, "finding the tenant cluster's control plane cloud formation stack outputs")

		o, s, err := cloudFormation.DescribeOutputsAndStatus(key.StackNameTCCP(&cr))
		if cloudformation.IsStackNotFound(err) {
			r.logger.Debugf(ctx, "did not find the tenant cluster's control plane cloud formation stack outputs")
			r.logger.Debugf(ctx, "the tenant cluster's control plane cloud formation stack does not exist")
			r.logger.Debugf(ctx, "canceling resource")
			return nil

		} else if cloudformation.IsOutputsNotAccessible(err) {
			r.logger.Debugf(ctx, "did not find the tenant cluster's control plane cloud formation stack outputs")
			r.logger.Debugf(ctx, "the tenant cluster's control plane cloud formation stack output values are not accessible due to stack status %#q", s)
			r.logger.Debugf(ctx, "canceling resource")
			cc.Status.TenantCluster.TCCP.IsTransitioning = true
			return nil

		} else if err != nil {
			return microerror.Mask(err)
		}

		outputs = o

		r.logger.Debugf(ctx, "found the tenant cluster's control plane cloud formation stack outputs")
	}

	if r.route53Enabled {
		{
			v, err := cloudFormation.GetOutputValue(outputs, APIServerPublicLoadBalancerKey)
			// migration code to dont throw error  when the old CF Stack dont yet have the new output value
			// TODO https://github.com/giantswarm/giantswarm/issues/13851
			// Related: https://github.com/giantswarm/giantswarm/issues/10139
			// after migration we can remove the check for IsOutputNotFound
			if cloudformation.IsOutputNotFound(err) {
				r.logger.Debugf(ctx, "did not find the tenant cluster's control plane APIServerPublicLoadBalancer output")
			} else {
				if err != nil {
					return microerror.Mask(err)
				}
				cc.Status.TenantCluster.DNS.APIPublicLoadBalancer = v
			}
		}

		{
			v, err := cloudFormation.GetOutputValue(outputs, HostedZoneID)
			if err != nil {
				return microerror.Mask(err)
			}
			cc.Status.TenantCluster.DNS.HostedZoneID = v
		}

		{
			v, err := cloudFormation.GetOutputValue(outputs, HostedZoneNameServersKey)
			if err != nil {
				return microerror.Mask(err)
			}
			cc.Status.TenantCluster.DNS.HostedZoneNameServers = v
		}

		{
			v, err := cloudFormation.GetOutputValue(outputs, InternalHostedZoneID)
			// We do not throw error when the TC does not
			// have internal hosted zone as it is not a strict requirement.
			//
			if cloudformation.IsOutputNotFound(err) {
				r.logger.Debugf(ctx, "did not find the tenant cluster's control plane internalHostedZoneID output")
			} else {
				if err != nil {
					return microerror.Mask(err)
				}
				cc.Status.TenantCluster.DNS.InternalHostedZoneID = v
			}
		}

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
