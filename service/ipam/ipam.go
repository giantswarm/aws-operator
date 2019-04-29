package ipam

import (
	"context"
	"fmt"
	"net"
	"sync"

	ipamlib "github.com/giantswarm/ipam"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
)

type Config struct {
	Logger micrologger.Logger
}

type ipam struct {
	logger micrologger.Logger
	mutex  *sync.Mutex
}

func New(config Config) (NetworkAllocator, error) {
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	r := &ipam{
		logger: config.Logger,
		mutex:  &sync.Mutex{},
	}

	return r, nil
}

func (i *ipam) Allocate(ctx context.Context, fullRange net.IPNet, netSize net.IPMask, callbacks AllocationCallbacks) (net.IPNet, error) {
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
		reservedSubnets, err = callbacks.GetReservedNetworks()
		if err != nil {
			return net.IPNet{}, microerror.Mask(err)
		}
	}

	var subnet net.IPNet
	{
		i.logger.LogCtx(ctx, "level", "debug", "message", "finding free subnet")

		subnet, err = ipamlib.Free(fullRange, netSize, reservedSubnets)
		if err != nil {
			return net.IPNet{}, microerror.Maskf(err, "networkRange: %s, allocatedSubnetMask: %s, reservedSubnets: %#v", fullRange.String(), netSize.String(), reservedSubnets)
		}

		i.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("found free subnet %#q", subnet.String()))
	}

	{
		i.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("persisting allocation candidate: %q", subnet.String()))

		err = callbacks.PersistAllocatedNetwork(subnet)
		if err != nil {
			return net.IPNet{}, microerror.Mask(err)
		}

		i.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("persisted allocation candidate: %q", subnet.String()))
	}

	return subnet, nil
}
