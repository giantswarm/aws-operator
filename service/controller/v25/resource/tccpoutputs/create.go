package tccpoutputs

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/giantswarm/microerror"

	cf "github.com/giantswarm/aws-operator/service/controller/v25/cloudformation"
	"github.com/giantswarm/aws-operator/service/controller/v25/controllercontext"
	"github.com/giantswarm/aws-operator/service/controller/v25/key"
)

func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	cr, err := key.ToCustomObject(obj)
	if err != nil {
		return microerror.Mask(err)
	}
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	var cloudFormation *cf.CloudFormation
	{
		c := cf.Config{
			Client: cc.Client.TenantCluster.AWS.CloudFormation,
		}

		cloudFormation, err = cf.New(c)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	var outputs []*cloudformation.Output
	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "finding the tenant cluster cloud formation stack outputs")

		o, s, err := cloudFormation.DescribeOutputsAndStatus(key.MainGuestStackName(cr))
		if cf.IsStackNotFound(err) {
			r.logger.LogCtx(ctx, "level", "debug", "message", "did not find the tenant cluster cloud formation stack outputs")
			r.logger.LogCtx(ctx, "level", "debug", "message", "the tenant cluster cloud formation stack does not exist")
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
			return nil

		} else if cf.IsOutputsNotAccessible(err) {
			r.logger.LogCtx(ctx, "level", "debug", "message", "did not find the tenant cluster cloud formation stack outputs")
			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("the tenant cluster main cloud formation stack output values are not accessible due to stack status %#q", s))
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
			return nil

		} else if err != nil {
			return microerror.Mask(err)
		}

		outputs = o

		r.logger.LogCtx(ctx, "level", "debug", "message", "found the tenant cluster cloud formation stack outputs")
	}

	if r.route53Enabled {
		v, err := cloudFormation.GetOutputValue(outputs, key.HostedZoneNameServers)
		if err != nil {
			return microerror.Mask(err)
		}
		cc.Status.TenantCluster.HostedZoneNameServers = v
	}

	{
		v, err := cloudFormation.GetOutputValue(outputs, key.VPCPeeringConnectionIDKey)
		if cf.IsOutputNotFound(err) {
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
			cc.Status.TenantCluster.VPCPeeringConnectionID = v
		} else if err != nil {
			return microerror.Mask(err)
		} else {
			cc.Status.TenantCluster.VPCPeeringConnectionID = v
		}
	}

	{
		v, err := cloudFormation.GetOutputValue(outputs, key.WorkerASGKey)
		if err != nil {
			return microerror.Mask(err)
		}
		cc.Status.TenantCluster.TCCP.ASG.Name = v
	}

	return nil
}

func searchPeeringConnectionID(client EC2, id string) (string, error) {
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
						aws.String(id),
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
