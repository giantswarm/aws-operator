package ipam

import (
	"context"
	"fmt"
	"net"

	"github.com/giantswarm/ipam"
	"github.com/giantswarm/microerror"
	"k8s.io/apimachinery/pkg/api/meta"

	"github.com/giantswarm/aws-operator/service/internal/locker"
)

// EnsureCreated allocates tenant cluster network segments. It gathers existing
// subnets from existing system resources like VPCs and Cluster CRs.
func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	m, err := meta.Accessor(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	{
		r.logger.LogCtx(ctx, "level", "debug", "message", "acquiring lock for IPAM")
		err := r.locker.Lock(ctx)
		if locker.IsAlreadyExists(err) {
			// In case the lock already exists we stop here and try again during
			// the next reconciliation loop because another process is already
			// trying to allocate subnets.
			r.logger.LogCtx(ctx, "level", "debug", "message", "lock for IPAM is already acquired")
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
			return nil
		} else if err != nil {
			return microerror.Mask(err)
		} else {
			r.logger.LogCtx(ctx, "level", "debug", "message", "acquired lock for IPAM")
		}

		defer func() {
			r.logger.LogCtx(ctx, "level", "debug", "message", "releasing lock for IPAM")
			err := r.locker.Unlock(ctx)
			if locker.IsNotFound(err) {
				r.logger.LogCtx(ctx, "level", "debug", "message", "lock for IPAM is already released")
			} else if err != nil {
				r.logger.LogCtx(ctx, "level", "error", "message", "failed to release lock for IPAM", "stack", fmt.Sprintf("%#v", err))
			} else {
				r.logger.LogCtx(ctx, "level", "debug", "message", "released lock for IPAM")
			}
		}()
	}

	{
		proceed, err := r.checker.Check(ctx, m.GetNamespace(), m.GetName())
		if err != nil {
			return microerror.Mask(err)
		}

		if !proceed {
			r.logger.LogCtx(ctx, "level", "debug", "message", "subnet already allocated")
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
			return nil
		}
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
