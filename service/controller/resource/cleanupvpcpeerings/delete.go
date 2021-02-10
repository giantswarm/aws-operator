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
				vpcPeeringConnectionIDs = append(vpcPeeringConnectionIDs, s.VpcPeeringConnectionId)
			}
		}

		r.logger.Debugf(ctx, "found %d vpc peering connections for vpc %s", len(vpcPeeringConnectionIDs), cl.Status.Provider.Network.VPCID)
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
