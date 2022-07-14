package ipam

import (
	"net"
	"reflect"

	"github.com/giantswarm/k8sclient/v7/pkg/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"

	"github.com/giantswarm/aws-operator/v12/service/internal/locker"
)

const (
	Name = "ipam"
)

const (
	// minAllocatedSubnetMaskBits is the maximum size of guest subnet i.e.
	// smaller number here -> larger subnet per guest cluster. For now anything
	// under 16 doesn't make sense in here.
	minAllocatedSubnetMaskBits = 16
)

type Config struct {
	Checker   Checker
	Collector Collector
	K8sClient k8sclient.Interface
	Locker    locker.Interface
	Logger    micrologger.Logger
	Persister Persister

	AllocatedSubnetMaskBits int
	NetworkRange            net.IPNet
	PrivateSubnetMaskBits   int
	PublicSubnetMaskBits    int
}

type Resource struct {
	checker   Checker
	collector Collector
	k8sClient k8sclient.Interface
	locker    locker.Interface
	logger    micrologger.Logger
	persister Persister

	allocatedSubnetMask net.IPMask
	networkRange        net.IPNet
}

func New(config Config) (*Resource, error) {
	if config.Checker == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Checker must not be empty", config)
	}
	if config.Collector == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Collector must not be empty", config)
	}
	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.K8sClient must not be empty", config)
	}
	if config.Locker == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Locker must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}
	if config.Persister == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Persister must not be empty", config)
	}

	if config.AllocatedSubnetMaskBits < minAllocatedSubnetMaskBits {
		return nil, microerror.Maskf(invalidConfigError, "%T.AllocatedSubnetMaskBits (%d) must not be smaller than %d", config, config.AllocatedSubnetMaskBits, minAllocatedSubnetMaskBits)
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
		checker:   config.Checker,
		collector: config.Collector,
		k8sClient: config.K8sClient,
		locker:    config.Locker,
		logger:    config.Logger,
		persister: config.Persister,

		allocatedSubnetMask: net.CIDRMask(config.AllocatedSubnetMaskBits, 32),
		networkRange:        config.NetworkRange,
	}

	return r, nil
}

func (r *Resource) Name() string {
	return Name
}
