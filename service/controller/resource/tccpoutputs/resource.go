package tccpoutputs

import (
	"context"

	infrastructurev1alpha2 "github.com/giantswarm/apiextensions/v3/pkg/apis/infrastructure/v1alpha2"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
)

const (
	Name = "tccpoutputs"
)

type Config struct {
	Logger micrologger.Logger

	Route53Enabled bool
	ToClusterFunc  func(ctx context.Context, v interface{}) (infrastructurev1alpha2.AWSCluster, error)
}

// Resource implements an operatorkit resource and provides a mechanism to fetch
// information from Cloud Formation stack outputs of the Tenant Cluster Control
// Plane stack.
//
// The TCCP manages the VPC Peering Connection. The peering connection ID is
// added to the controller context and used in the CPF stack.
//
type Resource struct {
	logger micrologger.Logger

	route53Enabled bool
	toClusterFunc  func(ctx context.Context, v interface{}) (infrastructurev1alpha2.AWSCluster, error)
}

func New(config Config) (*Resource, error) {
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}
	if config.ToClusterFunc == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.ToClusterFunc must not be empty", config)
	}

	r := &Resource{
		logger: config.Logger,

		route53Enabled: config.Route53Enabled,
		toClusterFunc:  config.ToClusterFunc,
	}

	return r, nil
}

func (r *Resource) Name() string {
	return Name
}
