package lifecycle

import (
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"

	awsclient "github.com/giantswarm/aws-operator/client/aws"
)

const (
	Name = "lifecyclev6"
)

type ResourceConfig struct {
	Clients awsclient.Clients
	Logger  micrologger.Logger
}

type Resource struct {
	clients awsclient.Clients
	logger  micrologger.Logger
}

func NewResource(config ResourceConfig) (*Resource, error) {
	if config.Clients.CloudFormation == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Clients.CloudFormation must not be empty", config)
	}
	if config.Clients.EC2 == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Clients.EC2 must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	newResource := &Resource{
		clients: config.Clients,
		logger: config.Logger.With(
			"resource", Name,
		),
	}

	return newResource, nil
}

func (r *Resource) Name() string {
	return Name
}
