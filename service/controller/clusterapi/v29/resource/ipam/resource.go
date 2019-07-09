package ipam

import (
	"net"
	"reflect"

	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"sigs.k8s.io/cluster-api/pkg/client/clientset_generated/clientset"

	"github.com/giantswarm/aws-operator/service/network"
)

const (
	Name = "ipamv29"
)

const (
	// minAllocatedSubnetMaskBits is the maximum size of guest subnet i.e.
	// smaller number here -> larger subnet per guest cluster. For now anything
	// under 16 doesn't make sense in here.
	minAllocatedSubnetMaskBits = 16
)

type Config struct {
	CMAClient        clientset.Interface
	G8sClient        versioned.Interface
	Logger           micrologger.Logger
	NetworkAllocator network.Allocator

	AllocatedSubnetMaskBits int
	AvailabilityZones       []string
	NetworkRange            net.IPNet
	PrivateSubnetMaskBits   int
	PublicSubnetMaskBits    int
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

	if config.AllocatedSubnetMaskBits < minAllocatedSubnetMaskBits {
		return nil, microerror.Maskf(invalidConfigError, "%T.AllocatedSubnetMaskBits (%d) must not be smaller than %d", config, config.AllocatedSubnetMaskBits, minAllocatedSubnetMaskBits)
	}
	if len(config.AvailabilityZones) == 0 {
		return nil, microerror.Maskf(invalidConfigError, "%T.AvailabilityZones must not be empty", config)
	}
	if reflect.DeepEqual(config.NetworkRange, net.IPNet{}) {
		return nil, microerror.Maskf(invalidConfigError, "%T.NetworkRange must not be empty", config)
	}
	if config.PrivateSubnetMaskBits <= config.AllocatedSubnetMaskBits {
		return nil, microerror.Maskf(invalidConfigError, "%T.PrivateSubnetMaskBits (%d) must not be smaller or equal than %T.AllocatedSubnetMaskBits (%d)", config, config.PrivateSubnetMaskBits, config, config.AllocatedSubnetMaskBits)
	}
	if config.PublicSubnetMaskBits <= config.AllocatedSubnetMaskBits {
		return nil, microerror.Maskf(invalidConfigError, "%T.PublicSubnetMaskBits (%d) must not be smaller or equal than %T.AllocatedSubnetMaskBits (%d)", config, config.PublicSubnetMaskBits, config, config.AllocatedSubnetMaskBits)
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
