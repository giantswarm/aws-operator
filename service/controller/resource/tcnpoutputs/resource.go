package tcnpoutputs

import (
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
)

const (
	Name = "tcnpoutputs"
)

type Config struct {
	Logger micrologger.Logger
}

// Resource implements an operatorkit resource and provides a mechanism to fetch
// information from Cloud Formation stack outputs of the Tenant Cluster Node
// Pool stack.
//
// The TCNP manages the node pools upon MachineDeployment CRs. For instance the
// TCNP stack contains the AWS ASG of the node pool and certain stack outputs
// which this resource collects and puts into the controller context.
type Resource struct {
	logger micrologger.Logger
}

func New(config Config) (*Resource, error) {
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	r := &Resource{
		logger: config.Logger,
	}

	return r, nil
}

func (r *Resource) Name() string {
	return Name
}
