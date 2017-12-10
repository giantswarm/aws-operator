package ipam

import (
	"context"
	"net"
	"time"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/microstorage"
)

const (
	IPAMSubnetStorageKey       = "/ipam/subnet"
	IPAMSubnetStorageKeyFormat = "/ipam/subnet/%s"
)

// Config represents the configuration used to create a new ipam service.
type Config struct {
	// Dependencies.
	Logger  micrologger.Logger
	Storage microstorage.Storage

	// Settings.
	// Network is the network in which all returned subnets should exist.
	Network *net.IPNet
}

// DefaultConfig provides a default configuration to create a new ipam service
// by best effort.
func DefaultConfig() Config {
	return Config{
		// Dependencies.
		Logger:  nil, // Required.
		Storage: nil, // Required.

		// Settings.
		Network: nil, // Required.
	}
}

// New creates a new configured ipam service.
func New(config Config) (*Service, error) {
	// Dependencies.
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "logger must not be empty")
	}
	if config.Storage == nil {
		return nil, microerror.Maskf(invalidConfigError, "storage must not be empty")
	}

	// Settings.
	if config.Network == nil {
		return nil, microerror.Maskf(invalidConfigError, "network must not be empty")
	}

	newService := &Service{
		// Dependencies.
		logger:  config.Logger,
		storage: config.Storage,

		// Settings.
		network: *config.Network,
	}

	return newService, nil
}

type Service struct {
	// Dependencies.
	logger  micrologger.Logger
	storage microstorage.Storage

	// Settings.
	network net.IPNet
}

// listSubnets retrieves the stored subnets from storage and returns them.
func (s *Service) listSubnets(ctx context.Context) ([]net.IPNet, error) {
	s.logger.Log("info", "listing subnets")

	k, err := microstorage.NewK(IPAMSubnetStorageKey)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	kvs, err := s.storage.List(ctx, k)
	if err != nil && !microstorage.IsNotFound(err) {
		return nil, microerror.Mask(err)
	}

	existingSubnets := []net.IPNet{}
	for _, kv := range kvs {
		// Storage returns the relative key with List, not the values.
		// Instead of then requesting each value, we revert the key to a valid
		// CIDR string.
		existingSubnetString := decodeRelativeKey(kv.Val())

		_, existingSubnet, err := net.ParseCIDR(existingSubnetString)
		if err != nil {
			return nil, microerror.Mask(err)
		}
		existingSubnets = append(existingSubnets, *existingSubnet)
	}

	subnetCounter.Set(float64(len(existingSubnets)))

	return existingSubnets, nil
}

// NewSubnet returns the next available subnet, of the configured size,
// from the configured network.
func (s *Service) NewSubnet(mask net.IPMask) (net.IPNet, error) {
	s.logger.Log("info", "creating new subnet")
	defer updateMetrics("create", time.Now())

	ctx := context.Background()

	existingSubnets, err := s.listSubnets(ctx)
	if err != nil {
		return net.IPNet{}, microerror.Mask(err)
	}

	s.logger.Log("info", "computing next subnet")
	subnet, err := Free(s.network, mask, existingSubnets)
	if err != nil {
		return net.IPNet{}, microerror.Mask(err)
	}

	s.logger.Log("info", "putting subnet", "subnet", subnet.String())
	kv, err := microstorage.NewKV(encodeKey(subnet), subnet.String())
	if err != nil {
		return net.IPNet{}, microerror.Mask(err)
	}
	if err := s.storage.Put(ctx, kv); err != nil {
		return net.IPNet{}, microerror.Mask(err)
	}

	return subnet, nil
}

// DeleteSubnet deletes the given subnet from IPAM storage,
// meaning it can be given out again.
func (s *Service) DeleteSubnet(subnet net.IPNet) error {
	s.logger.Log("info", "deleting subnet", "subnet", subnet.String())
	defer updateMetrics("delete", time.Now())

	ctx := context.Background()

	k, err := microstorage.NewK(encodeKey(subnet))
	if err != nil {
		return microerror.Mask(err)
	}
	if err := s.storage.Delete(ctx, k); err != nil {
		return microerror.Mask(err)
	}

	return nil
}
