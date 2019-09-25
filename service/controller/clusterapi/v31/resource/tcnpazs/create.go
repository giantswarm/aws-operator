package tcnpazs

import (
	"context"
	"net"
	"sort"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/giantswarm/ipam"
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/pkg/awstags"
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v30/controllercontext"
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v30/key"
)

func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	cr, err := key.ToMachineDeployment(obj)
	if err != nil {
		return microerror.Mask(err)
	}
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	// We need to cancel the resource early in case the ipam resource did not yet
	// allocate a subnet for the node pool.
	if key.MachineDeploymentSubnet(cr) == "" {
		r.logger.LogCtx(ctx, "level", "debug", "message", "cannot collect private subnets for availability zones")
		r.logger.LogCtx(ctx, "level", "debug", "message", "node pool subnet not yet allocated")
		r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")

		return nil
	}

	// Split the node pool subnet by the number of availability zones for further
	// mapping below.
	var subnets []net.IPNet
	{
		_, netip, err := net.ParseCIDR(key.MachineDeploymentSubnet(cr))
		if err != nil {
			return microerror.Mask(err)
		}

		subnets, err = ipam.Split(*netip, uint(len(key.MachineDeploymentAvailabilityZones(cr))))
		if err != nil {
			return microerror.Mask(err)
		}
	}

	// Ensure a copy of the list of availability zones. This is absolutely
	// critical for the mapping between availability zones and their split
	// private subnets.
	var azs []string
	{
		for _, az := range key.MachineDeploymentAvailabilityZones(cr) {
			azs = append(azs, az)
		}

		sort.Strings(azs)
	}

	// Put the mapping between availability zones and their split subnets into the
	// controller context spec in a deterministic way.
	{
		var list []controllercontext.ContextSpecTenantClusterTCNPAvailabilityZone

		for i, az := range azs {
			ng := natGatewayForAvailabilityZone(cc.Status.TenantCluster.TCCP.NATGateways, az)
			if ng == nil {
				r.logger.LogCtx(ctx, "level", "debug", "message", "nat gateway information not yet available")
				r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")

				return nil
			}

			item := controllercontext.ContextSpecTenantClusterTCNPAvailabilityZone{
				Name: az,
				NATGateway: controllercontext.ContextSpecTenantClusterTCNPAvailabilityZoneNATGateway{
					ID: *ng.NatGatewayId,
				},
				Subnet: controllercontext.ContextSpecTenantClusterTCNPAvailabilityZoneSubnet{
					Private: controllercontext.ContextSpecTenantClusterTCNPAvailabilityZoneSubnetPrivate{
						CIDR: subnets[i],
					},
				},
			}

			list = append(list, item)
		}

		cc.Spec.TenantCluster.TCNP.AvailabilityZones = list
	}

	return nil
}

func natGatewayForAvailabilityZone(natGateways []*ec2.NatGateway, availabilityZone string) *ec2.NatGateway {
	for _, ng := range natGateways {
		if awstags.ValueForKey(ng.Tags, key.TagStack) != key.StackTCCP {
			continue
		}
		if awstags.ValueForKey(ng.Tags, key.TagAvailabilityZone) != availabilityZone {
			continue
		}

		return ng
	}

	return nil
}
