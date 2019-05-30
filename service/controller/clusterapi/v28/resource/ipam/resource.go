package ipam

import (
	"net"
	"reflect"

	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/aws-operator/service/network"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"sigs.k8s.io/cluster-api/pkg/client/clientset_generated/clientset"
)

const (
	Name = "ipamv27"
)

type Config struct {
	CMAClient        clientset.Interface
	G8sClient        versioned.Interface
	Logger           micrologger.Logger
	NetworkAllocator network.Allocator

	AllocatedSubnetMaskBits int
	AvailabilityZones       []string
	NetworkRange            net.IPNet
}

type Resource struct {
	cmaClient        clientset.Interface
	g8sClient        versioned.Interface
	logger           micrologger.Logger
	networkAllocator network.Allocator

	allocatedSubnetMask net.IPMask
	availabilityZones   []string
	networkRange        net.IPNet
}

func New(config Config) (*Resource, error) {
	if config.CMAClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.CMAClient must not be empty", config)
	}
	if config.G8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.G8sClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}
	if config.NetworkAllocator == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.NetworkAllocator must not be empty", config)
	}

	if len(config.AvailabilityZones) == 0 {
		return nil, microerror.Maskf(invalidConfigError, "%T.AvailabilityZones must not be empty", config)
	}
	if reflect.DeepEqual(config.NetworkRange, net.IPNet{}) {
		return nil, microerror.Maskf(invalidConfigError, "%T.NetworkRange must not be empty", config)
	}

	r := &Resource{
		cmaClient:        config.CMAClient,
		g8sClient:        config.G8sClient,
		logger:           config.Logger,
		networkAllocator: config.NetworkAllocator,

		allocatedSubnetMask: net.CIDRMask(config.AllocatedSubnetMaskBits, 32),
		availabilityZones:   config.AvailabilityZones,
		networkRange:        config.NetworkRange,
	}

	return r, nil
}

func (r *Resource) Name() string {
	return Name
}
