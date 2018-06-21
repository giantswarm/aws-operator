package provider

import (
	"github.com/giantswarm/e2e-harness/pkg/framework"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
)

type KVMConfig struct {
	HostFramework *framework.Host
	Logger        micrologger.Logger

	ClusterID string
}

type KVM struct {
	hostFramework *framework.Host
	logger        micrologger.Logger

	clusterID string
}

func NewKVM(config KVMConfig) (*KVM, error) {
	if config.HostFramework == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.HostFramework must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	if config.ClusterID == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.ClusterID must not be empty", config)
	}

	a := &KVM{
		hostFramework: config.HostFramework,
		logger:        config.Logger,

		clusterID: config.ClusterID,
	}

	return a, nil
}

func (a *KVM) RebootMaster() error {
	// TOOD
	return nil
}

func (a *KVM) ReplaceMaster() error {
	// TOOD
	return nil
}
