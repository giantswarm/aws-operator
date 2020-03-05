// Package tccpazs implements a resource to gather all distinct availability
// zones for a tenant cluster.
package tccpazs

import (
	infrastructurev1alpha2 "github.com/giantswarm/apiextensions/pkg/apis/infrastructure/v1alpha2"
	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
)

const (
	Name = "tccpazs"
)

type Config struct {
	G8sClient     versioned.Interface
	Logger        micrologger.Logger
	ToClusterFunc func(v interface{}) (infrastructurev1alpha2.AWSCluster, error)

	CIDRBlockAWSCNI string
}

type Resource struct {
	g8sClient     versioned.Interface
	logger        micrologger.Logger
	toClusterFunc func(v interface{}) (infrastructurev1alpha2.AWSCluster, error)

	cidrBlockAWSCNI string
}

func New(config Config) (*Resource, error) {
	if config.G8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.G8sClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}
	if config.ToClusterFunc == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.ToClusterFunc must not be empty", config)
	}

	if config.CIDRBlockAWSCNI == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.CIDRBlockAWSCNI must not be empty", config)
	}

	r := &Resource{
		g8sClient:     config.G8sClient,
		logger:        config.Logger,
		toClusterFunc: config.ToClusterFunc,

		cidrBlockAWSCNI: config.CIDRBlockAWSCNI,
	}

	return r, nil
}

func (r *Resource) Name() string {
	return Name
}
