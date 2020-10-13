package cleanupenis

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/v2/pkg/controller/context/finalizerskeptcontext"

	"github.com/giantswarm/aws-operator/service/controller/controllercontext"
)

func (r *Resource) EnsureDeleted(ctx context.Context, obj interface{}) error {
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	// We need to fetch all subnets of the Tenant Cluster in order to find all
	// relevant ENIs.
	var values []*string
	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "finding all subnets")

		i := &ec2.DescribeSubnetsInput{
			Filters: []*ec2.Filter{
				{
					Name: aws.String("vpc-id"),
					Values: []*string{
						aws.String(cc.Status.TenantCluster.TCCP.VPC.ID),
					},
				},
			},
		}

		o, err := cc.Client.TenantCluster.AWS.EC2.DescribeSubnets(i)
		if err != nil {
			return microerror.Mask(err)
		}

		for _, s := range o.Subnets {
			values = append(values, s.SubnetId)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("found %d subnets", len(values)))
	}

	var enis []*ec2.NetworkInterface
	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "finding network interfaces")

		i := &ec2.DescribeNetworkInterfacesInput{
			Filters: []*ec2.Filter{
				{
					Name:   aws.String("subnet-id"),
					Values: values,
				},
			},
		}

		o, err := cc.Client.TenantCluster.AWS.EC2.DescribeNetworkInterfaces(i)
		if err != nil {
			return microerror.Mask(err)
		}

		enis = o.NetworkInterfaces

		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("found %d network interfaces", len(enis)))
	}

	// We want to cleanup network interfaces. We need to check which ENIs are
	// attached and which are detached. When a network interface is still
	// attached, its status is in-use. When it is already detached, its status
	// is available. See e.g. the CLI docs below.
	//
	//     https://docs.aws.amazon.com/cli/latest/reference/ec2/wait/network-interface-available.html
	//
	var attached []*ec2.NetworkInterface
	var detached []*ec2.NetworkInterface
	var transitioning []*ec2.NetworkInterface
	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "grouping network interfaces")

		for _, eni := range enis {
			switch *eni.Status {
			case ec2.NetworkInterfaceStatusInUse:
				attached = append(attached, eni)
			case ec2.NetworkInterfaceStatusAvailable:
				detached = append(detached, eni)
			default:
				transitioning = append(transitioning, eni)
			}
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("found %d attached network interfaces", len(attached)))
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("found %d detached network interfaces", len(detached)))
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("found %d transitioning network interfaces", len(transitioning)))
	}

	// For all the detached ENIs we try to delete them. This is the cleanup
	// mechanism we want to ensure due to some insuficiencies in the AWS CNI.
	for _, eni := range detached {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("deleting detached network interface %#q", *eni.NetworkInterfaceId))

		i := &ec2.DeleteNetworkInterfaceInput{
			NetworkInterfaceId: eni.NetworkInterfaceId,
		}

		_, err := cc.Client.TenantCluster.AWS.EC2.DeleteNetworkInterface(i)
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("deleted detached network interface %#q", *eni.NetworkInterfaceId))
	}

	// For all attached ENIs we just keep finalizers and try cleaning them up
	// again during the next reconciliation loop. The same applies for any
	// network interfaces transitioning between states. Transitioning states
	// indicate that e.g. ENIs are currently being detached from their
	// respective instances during Tenant Cluster deletion.
	if len(attached) > 0 || len(transitioning) > 0 {
		r.logger.LogCtx(ctx, "level", "debug", "message", "found network interfaces which are not yet detached")
		r.logger.LogCtx(ctx, "level", "debug", "message", "keeping finalizers")
		finalizerskeptcontext.SetKept(ctx)
	}

	return nil
}
