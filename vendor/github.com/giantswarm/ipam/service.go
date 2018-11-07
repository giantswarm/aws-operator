package ipam

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/microstorage"
)

const (
	ipamSubnetStorageKey = "/ipam/subnet"
)

// Config represents the configuration used to create a new ipam service.
type Config struct {
	Logger  micrologger.Logger
	Storage microstorage.Storage

	// Network is the network in which all returned subnets should exist.
	Network *net.IPNet
	// AllocatedSubnets is a list of subnets, contained by `Network`,
	// that have already been allocated outside of IPAM control.
	// Any subnets created by the IPAM service will not overlap with these subnets.
	AllocatedSubnets []net.IPNet
}

// New creates a new configured ipam service.
func New(config Config) (*Service, error) {
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "logger must not be empty")
	}
	if config.Storage == nil {
		return nil, microerror.Maskf(invalidConfigError, "storage must not be empty")
	}

	if config.Network == nil {
		return nil, microerror.Maskf(invalidConfigError, "network must not be empty")
	}
	for _, allocatedSubnet := range config.AllocatedSubnets {
		ipRange := newIPRange(allocatedSubnet)
		if !(config.Network.Contains(ipRange.start) && config.Network.Contains(ipRange.end)) {
			return nil, microerror.Maskf(
				invalidConfigError,
				"allocated subnet (%v) must be contained by network (%v)",
				allocatedSubnet.String(),
				config.Network.String(),
			)
		}
	}

	newService := &Service{
		logger:  config.Logger,
		storage: config.Storage,

		network:          *config.Network,
		allocatedSubnets: config.AllocatedSubnets,
	}

	return newService, nil
}

type Service struct {
	logger  micrologger.Logger
	storage microstorage.Storage

	network          net.IPNet
	allocatedSubnets []net.IPNet
}

// listSubnets retrieves the stored subnets from storage and returns them.
func (s *Service) listSubnets(ctx context.Context) ([]net.IPNet, error) {
	s.logger.LogCtx(ctx, "level", "info", "message", "listing subnets")

	k, err := microstorage.NewK(ipamSubnetStorageKey)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	kvs, err := s.storage.List(ctx, k)
	if err != nil && !microstorage.IsNotFound(err) {
		return nil, microerror.Mask(err)
	}

	existingSubnets := []net.IPNet{}
	for _, kv := range kvs {
		existingSubnetString := decodeKey(kv.Key())

		_, existingSubnet, err := net.ParseCIDR(existingSubnetString)
		if err != nil {
			return nil, microerror.Mask(err)
		}
		existingSubnets = append(existingSubnets, *existingSubnet)
	}

	subnetCounter.Set(float64(len(existingSubnets)))

	return existingSubnets, nil
}

// CreateSubnet returns the next available subnet, of the configured size,
// from the configured network.
func (s *Service) CreateSubnet(ctx context.Context, mask net.IPMask, annotation string, reserved []net.IPNet) (net.IPNet, error) {
	s.logger.LogCtx(ctx, "level", "debug", "message", "creating subnet")
	defer updateMetrics("create", time.Now())

	existingSubnets, err := s.listSubnets(ctx)
	if err != nil {
		return net.IPNet{}, microerror.Mask(err)
	}

	existingSubnets = append(existingSubnets, reserved...)
	existingSubnets = append(existingSubnets, s.allocatedSubnets...)
	existingSubnets = CanonicalizeSubnets(s.network, existingSubnets)

	subnet, err := Free(s.network, mask, existingSubnets)
	if err != nil {
		return net.IPNet{}, microerror.Mask(err)
	}

	kv, err := microstorage.NewKV(encodeKey(subnet), annotation)
	if err != nil {
		return net.IPNet{}, microerror.Mask(err)
	}
	if err := s.storage.Put(ctx, kv); err != nil {
		return net.IPNet{}, microerror.Mask(err)
	}

	s.logger.LogCtx(ctx, "level", "debug", "message", "created subnet")

	return subnet, nil
}

// DeleteSubnet deletes the given subnet from IPAM storage,
// meaning it can be given out again.
func (s *Service) DeleteSubnet(ctx context.Context, subnet net.IPNet) error {
	s.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("deleting subnet %#q", subnet.String()))
	defer updateMetrics("delete", time.Now())

	k, err := microstorage.NewK(encodeKey(subnet))
	if err != nil {
		return microerror.Mask(err)
	}
	if err := s.storage.Delete(ctx, k); err != nil {
		return microerror.Mask(err)
	}

	s.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("deleted subnet %#q", subnet.String()))

	return nil
}
