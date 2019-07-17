package machinedeploymentazs

import (
	"context"
	"net"
	"sort"

	"github.com/giantswarm/ipam"
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v29/controllercontext"
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v29/key"
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
	if key.WorkerSubnet(cr) == "" {
		r.logger.LogCtx(ctx, "level", "debug", "message", "cannot collect private subnets for availability zones")
		r.logger.LogCtx(ctx, "level", "debug", "message", "node pool subnet not yet allocated")
		r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")

		return nil
	}

	// Split the node pool subnet by the number of availability zones for further
	// mapping below.
	var subnets []net.IPNet
	{
		_, netip, err := net.ParseCIDR(key.WorkerSubnet(cr))
		if err != nil {
			return microerror.Mask(err)
		}

		subnets, err = ipam.Split(*netip, uint(len(key.WorkerAvailabilityZones(cr))))
		if err != nil {
			return microerror.Mask(err)
		}
	}

	// Ensure a copy of the list of availability zones. This is absolutely
	// critical for the mapping between availability zones and their split
	// private subnets.
	var azs []string
	{
		for _, az := range key.WorkerAvailabilityZones(cr) {
			azs = append(azs, az)
		}

		sort.Strings(azs)
	}

	// Put the mapping between availability zones and their split subnets into the
	// controller context spec in a deterministic way.
	{
		var list []controllercontext.ContextSpecTenantClusterTCNPAvailabilityZone

		for i, az := range azs {
			item := controllercontext.ContextSpecTenantClusterTCNPAvailabilityZone{
				AvailabilityZone: az,
				PrivateSubnet:    subnets[i],
			}

			list = append(list, item)
		}

		cc.Spec.TenantCluster.TCNP.AvailabilityZones = list
	}

	return nil
}
