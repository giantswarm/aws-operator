package network

import (
	"context"
	"fmt"
	"net"

	"github.com/giantswarm/ipam"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"

	"github.com/giantswarm/aws-operator/service/internal/locker"
)

type Config struct {
	Locker locker.Interface
	Logger micrologger.Logger
}

type allocator struct {
	locker locker.Interface
	logger micrologger.Logger
}

func New(config Config) (Allocator, error) {
	if config.Locker == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Locker must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	r := &allocator{
		locker: config.Locker,
		logger: config.Logger,
	}

	return r, nil
}

func (i *allocator) Allocate(ctx context.Context, fullRange net.IPNet, netSize net.IPMask, callbacks AllocationCallbacks) (net.IPNet, error) {
	{
		i.logger.LogCtx(ctx, "level", "debug", "message", "acquiring lock for IPAM")
		err := i.locker.Lock(ctx)
		if locker.IsAlreadyExists(err) {
			i.logger.LogCtx(ctx, "level", "debug", "message", "lock for IPAM is already acquired")
		} else if err != nil {
			return net.IPNet{}, microerror.Mask(err)
		} else {
			i.logger.LogCtx(ctx, "level", "debug", "message", "acquired lock for IPAM")
		}

		defer func() {
			i.logger.LogCtx(ctx, "level", "debug", "message", "releasing lock for IPAM")
			err := i.locker.Unlock(ctx)
			if locker.IsNotFound(err) {
				i.logger.LogCtx(ctx, "level", "debug", "message", "lock for IPAM is already released")
			} else if err != nil {
				i.logger.LogCtx(ctx, "level", "error", "message", "failed to release lock for IPAM", "stack", fmt.Sprintf("%#v", err))
			} else {
				i.logger.LogCtx(ctx, "level", "debug", "message", "released lock for IPAM")
			}
		}()
	}

	var err error
	var reservedSubnets []net.IPNet
	{
		i.logger.LogCtx(ctx, "level", "debug", "message", "getting allocated subnets")

		reservedSubnets, err = callbacks.GetReservedNetworks(ctx)
		if err != nil {
			return net.IPNet{}, microerror.Mask(err)
		}

		i.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("got allocated subnets: %q", reservedSubnets))
	}

	var subnet net.IPNet
	{
		i.logger.LogCtx(ctx, "level", "debug", "message", "finding free subnet")

		subnet, err = ipam.Free(fullRange, netSize, reservedSubnets)
		if err != nil {
			return net.IPNet{}, microerror.Mask(err)
		}

		i.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("found free subnet %#q", subnet.String()))
	}

	{
		i.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("persisting allocation candidate: %q", subnet.String()))

		err = callbacks.PersistAllocatedNetwork(ctx, subnet)
		if err != nil {
			return net.IPNet{}, microerror.Mask(err)
		}

		i.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("persisted allocation candidate: %q", subnet.String()))
	}

	return subnet, nil
}
