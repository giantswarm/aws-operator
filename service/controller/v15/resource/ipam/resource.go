package ipam

import (
	"net"
	"reflect"

	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
)

const (
	Name = "ipamv15"
)

type Config struct {
	G8sClient versioned.Interface
	Logger    micrologger.Logger

	AllocatedSubnetMaskBits int
	NetworkRange            net.IPNet
}

type Resource struct {
	g8sClient versioned.Interface
	logger    micrologger.Logger

	allocatedSubnetMask net.IPMask
	networkRange        net.IPNet
}

func New(config Config) (*Resource, error) {
	if config.G8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.G8sClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}
	if reflect.DeepEqual(config.NetworkRange, net.IPNet{}) {
		return nil, microerror.Maskf(invalidConfigError, "%T.NetworkRange must not be empty", config)
	}

	newResource := &Resource{
		g8sClient: config.G8sClient,
		logger:    config.Logger,

		allocatedSubnetMask: net.CIDRMask(config.AllocatedSubnetMaskBits, 32),
		networkRange:        config.NetworkRange,
	}

	return newResource, nil
}

func (r *Resource) Name() string {
	return Name
}
