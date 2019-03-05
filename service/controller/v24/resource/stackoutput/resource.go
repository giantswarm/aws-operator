package stackoutput

import (
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
)

const (
	Name = "stackoutputv24"
)

type Config struct {
	Logger micrologger.Logger

	Route53Enabled bool
}

// Resource implements an operatorkit resource and provides a mechanism to fetch
// information from a cloud formation stack outputs of the Tenant Cluster Main
// stack.
//
// The TCM manages the VPC Peering Connection. The peering connection ID is
// added to the controller context and used in the CPF stack.
//
type Resource struct {
	logger micrologger.Logger

	route53Enabled bool
}

func New(config Config) (*Resource, error) {
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	r := &Resource{
		logger: config.Logger,

		route53Enabled: config.Route53Enabled,
	}

	return r, nil
}

func (r *Resource) Name() string {
	return Name
}
