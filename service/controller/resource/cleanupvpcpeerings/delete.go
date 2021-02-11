package cleanupvpcpeerings

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/service/controller/controllercontext"
	"github.com/giantswarm/aws-operator/service/controller/key"
)

func (r *Resource) EnsureDeleted(ctx context.Context, obj interface{}) error {
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	cl, err := key.ToCluster(ctx, obj)
	if err != nil {
		return microerror.Mask(err)
	}

	// Fetch all VPC peering connections of the Tenant Cluster VPC
	var vpcPeeringConnectionIDs []*string
	{
		r.logger.Debugf(ctx, "finding all vpc peering connections")

		// fetch all vpc peering connection via requester vpc filter
		{
			i := &ec2.DescribeVpcPeeringConnectionsInput{
				Filters: []*ec2.Filter{
					{
						Name: aws.String("requester-vpc-info.vpc-id"),
						Values: []*string{
							aws.String(cl.Status.Provider.Network.VPCID),
						},
					},
				},
			}

			o, err := cc.Client.TenantCluster.AWS.EC2.DescribeVpcPeeringConnections(i)
			if err != nil {
				return microerror.Mask(err)
			}

			for _, s := range o.VpcPeeringConnections {
				if isTCCPVPCPeering(s.Tags) {
					// skip deleting vpc peering connection from tccp CF stack
					// it will be deleted by the CF
					continue
				}
				if *s.Status.Code == vpcStatusDeleting || *s.Status.Code == vpcStatusDeleted {
					// ignore vpc peering connections that are already deleting or deleted
					continue
				}

				vpcPeeringConnectionIDs = append(vpcPeeringConnectionIDs, s.VpcPeeringConnectionId)
			}
		}
		// fetch all vpc peering connection via accepter vpc filter
		{
			i := &ec2.DescribeVpcPeeringConnectionsInput{
				Filters: []*ec2.Filter{
					{
						Name: aws.String("accepter-vpc-info.vpc-id"),
						Values: []*string{
							aws.String(cl.Status.Provider.Network.VPCID),
						},
					},
				},
			}

			o, err := cc.Client.TenantCluster.AWS.EC2.DescribeVpcPeeringConnections(i)
			if err != nil {
				return microerror.Mask(err)
			}

			for _, s := range o.VpcPeeringConnections {
				if isTCCPVPCPeering(s.Tags) {
					// skip deleting vpc peering connection from tccp CF stack
					// it will be deleted by the CF
					continue
				}
				if *s.Status.Code == vpcStatusDeleting || *s.Status.Code == vpcStatusDeleted {
					// ignore vpc peering connections that are already deleting or deleted
					continue
				}

				vpcPeeringConnectionIDs = append(vpcPeeringConnectionIDs, s.VpcPeeringConnectionId)
			}
		}

		r.logger.Debugf(ctx, "found %d vpc peering connections for %s", len(vpcPeeringConnectionIDs), cl.Status.Provider.Network.VPCID)
	}

	for _, id := range vpcPeeringConnectionIDs {
		r.logger.Debugf(ctx, "deleting vpc peering connection %#q", *id)

		i := &ec2.DeleteVpcPeeringConnectionInput{
			VpcPeeringConnectionId: id,
		}

		_, err := cc.Client.TenantCluster.AWS.EC2.DeleteVpcPeeringConnection(i)
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.Debugf(ctx, "deleted vpc peering connection %#q", *id)
	}
	return nil
}

func isTCCPVPCPeering(tags []*ec2.Tag) bool {
	for _, tag := range tags {
		if *tag.Key == key.TagStack && *tag.Value == key.StackTCCP {
			return true
		}
	}

	return false
}
