package drainer

import (
	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"
)

const (
	Name = "drainerv29"
)

type ResourceConfig struct {
	G8sClient     versioned.Interface
	Logger        micrologger.Logger
	ToClusterFunc func(v interface{}) (v1alpha1.Cluster, error)
}

type Resource struct {
	g8sClient     versioned.Interface
	logger        micrologger.Logger
	toClusterFunc func(v interface{}) (v1alpha1.Cluster, error)
}

func NewResource(config ResourceConfig) (*Resource, error) {
	if config.G8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.G8sClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}
	if config.ToClusterFunc == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.ToClusterFunc must not be empty", config)
	}

	r := &Resource{
		g8sClient:     config.G8sClient,
		logger:        config.Logger,
		toClusterFunc: config.ToClusterFunc,
	}

	return r, nil
}

func (r *Resource) Name() string {
	return Name
}
