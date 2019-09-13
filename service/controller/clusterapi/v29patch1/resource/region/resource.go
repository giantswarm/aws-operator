// Package region implements an operatorkit resource that addresses a problem
// where the tcnp resource would need to fetch the Cluster CR even though the
// MachineDeployment CR is reconciled. This is only because we need the AWS
// region to lookup S3 bucket names and EC2 image IDs and the like. So in order
// to free the tcnp resource from that hustle we implement a separate region
// resource which does the lookup and puts the region into the controller
// context. The controller context information are then simply used by the tcnp
// resource as this is our state of the art primitive for information
// distribution within a controller's reconciliation.
package region

import (
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"
)

const (
	Name = "regionv29"
)

type Config struct {
	Logger        micrologger.Logger
	ToClusterFunc func(v interface{}) (v1alpha1.Cluster, error)
}

type Resource struct {
	logger        micrologger.Logger
	toClusterFunc func(v interface{}) (v1alpha1.Cluster, error)
}

func New(config Config) (*Resource, error) {
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}
	if config.ToClusterFunc == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.ToClusterFunc must not be empty", config)
	}

	r := &Resource{
		logger:        config.Logger,
		toClusterFunc: config.ToClusterFunc,
	}

	return r, nil
}

func (r *Resource) Name() string {
	return Name
}
