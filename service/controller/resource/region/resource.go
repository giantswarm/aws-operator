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
	"context"

	infrastructurev1alpha2 "github.com/giantswarm/apiextensions/v3/pkg/apis/infrastructure/v1alpha2"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
)

const (
	Name = "region"
)

type Config struct {
	Logger        micrologger.Logger
	ToClusterFunc func(ctx context.Context, v interface{}) (infrastructurev1alpha2.AWSCluster, error)
}

type Resource struct {
	logger        micrologger.Logger
	toClusterFunc func(ctx context.Context, v interface{}) (infrastructurev1alpha2.AWSCluster, error)
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
