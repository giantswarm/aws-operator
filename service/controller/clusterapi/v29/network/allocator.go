package network

import (
	"context"
	"fmt"
	"net"
	"sync"

	"github.com/giantswarm/ipam"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
)

type Config struct {
	Logger micrologger.Logger
}

type allocator struct {
	logger micrologger.Logger
	mutex  *sync.Mutex
}

func New(config Config) (Allocator, error) {
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	r := &allocator{
		logger: config.Logger,
		mutex:  &sync.Mutex{},
	}

	return r, nil
}

func (i *allocator) Allocate(ctx context.Context, fullRange net.IPNet, netSize net.IPMask, callbacks AllocationCallbacks) (net.IPNet, error) {
	i.logger.LogCtx(ctx, "level", "debug", "message", "acquiring lock for IPAM")
	i.mutex.Lock()
	i.logger.LogCtx(ctx, "level", "debug", "message", "acquired lock for IPAM")

	defer func() {
		i.logger.LogCtx(ctx, "level", "debug", "message", "releasing lock for IPAM")
		i.mutex.Unlock()
		i.logger.LogCtx(ctx, "level", "debug", "message", "released lock for IPAM")
	}()

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
			return net.IPNet{}, microerror.Maskf(err, "networkRange: %s, allocatedSubnetMask: %s, reservedSubnets: %#v", fullRange.String(), netSize.String(), reservedSubnets)
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
