package ipam

import (
	"context"
	"fmt"
	"net"

	"github.com/giantswarm/ipam"
	"github.com/giantswarm/microerror"
	"k8s.io/apimachinery/pkg/api/meta"
)

// EnsureCreated allocates tenant cluster network segments. It gathers existing
// subnets from existing system resources like VPCs and Cluster CRs.
func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	m, err := meta.Accessor(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	proceed, err := r.checker.Check(ctx, m.GetNamespace(), m.GetName())
	if err != nil {
		return microerror.Mask(err)
	}

	if !proceed {
		r.logger.LogCtx(ctx, "level", "debug", "message", "subnet already allocated")
		return nil
	}

	var allocatedSubnets []net.IPNet
	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "finding allocated subnets")

		allocatedSubnets, err = r.collector.Collect(ctx)
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("found allocated subnets %#q", allocatedSubnets))
	}

	var freeSubnet net.IPNet
	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "finding free subnet")

		freeSubnet, err = ipam.Free(r.networkRange, r.allocatedSubnetMask, allocatedSubnets)
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("found free subnet %#q", freeSubnet))
	}

	{
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("allocating free subnet %#q", freeSubnet))

		err = r.persister.Persist(ctx, freeSubnet, m.GetNamespace(), m.GetName())
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("allocated free subnet %#q", freeSubnet))
	}

	return nil
}
