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

type IPAMAllocatorConfig struct {
	Logger micrologger.Logger
}

type IPAMAllocator struct {
	logger micrologger.Logger

	mutex *sync.Mutex
}

func NewIPAMAllocater(config IPAMAllocatorConfig) (*IPAMAllocator, error) {
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	a := &IPAMAllocator{
		logger: config.Logger,

		mutex: &sync.Mutex{},
	}

	return a, nil
}

func (a *IPAMAllocator) Allocate(ctx context.Context, fullRange net.IPNet, netSize net.IPMask, callbacks Callbacks) (net.IPNet, error) {
	a.logger.LogCtx(ctx, "level", "debug", "message", "acquiring lock for IPAM")
	a.mutex.Lock()
	a.logger.LogCtx(ctx, "level", "debug", "message", "acquired lock for IPAM")

	defer func() {
		a.logger.LogCtx(ctx, "level", "debug", "message", "releasing lock for IPAM")
		a.mutex.Unlock()
		a.logger.LogCtx(ctx, "level", "debug", "message", "released lock for IPAM")
	}()

	var err error
	var reservedSubnets []net.IPNet
	{
		a.logger.LogCtx(ctx, "level", "debug", "message", "getting allocated subnets")

		reservedSubnets, err = callbacks.Collect(ctx)
		if err != nil {
			return net.IPNet{}, microerror.Mask(err)
		}

		a.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("got allocated subnets: %q", reservedSubnets))
	}

	var subnet net.IPNet
	{
		a.logger.LogCtx(ctx, "level", "debug", "message", "finding free subnet")

		subnet, err = ipam.Free(fullRange, netSize, reservedSubnets)
		if err != nil {
			return net.IPNet{}, microerror.Maskf(err, "networkRange: %s, allocatedSubnetMask: %s, reservedSubnets: %#v", fullRange.String(), netSize.String(), reservedSubnets)
		}

		a.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("found free subnet %#q", subnet.String()))
	}

	{
		a.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("persisting allocation candidate: %q", subnet.String()))

		err = callbacks.Persist(ctx, subnet)
		if err != nil {
			return net.IPNet{}, microerror.Mask(err)
		}

		a.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("persisted allocation candidate: %q", subnet.String()))
	}

	return subnet, nil
}
