// Package tccpazs implements a resource to gather all distinct availability
// zones for a tenant cluster.
package tccpazs

import (
	"github.com/giantswarm/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
)

const (
	Name = "tccpazs"
)

type Config struct {
	K8sClient k8sclient.Interface
	Logger    micrologger.Logger

	CIDRBlockAWSCNI string
}

type Resource struct {
	k8sClient k8sclient.Interface
	logger    micrologger.Logger

	cidrBlockAWSCNI string
}

func New(config Config) (*Resource, error) {
	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.K8sClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	if config.CIDRBlockAWSCNI == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.CIDRBlockAWSCNI must not be empty", config)
	}

	r := &Resource{
		k8sClient: config.K8sClient,
		logger:    config.Logger,

		cidrBlockAWSCNI: config.CIDRBlockAWSCNI,
	}

	return r, nil
}

func (r *Resource) Name() string {
	return Name
}
