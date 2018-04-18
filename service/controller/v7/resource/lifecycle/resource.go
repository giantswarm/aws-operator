package lifecycle

import (
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"

	awsclient "github.com/giantswarm/aws-operator/client/aws"
	cloudformationservice "github.com/giantswarm/aws-operator/service/controller/v7/cloudformation"
)

const (
	Name = "lifecyclev7"
)

type ResourceConfig struct {
	Clients awsclient.Clients
	Logger  micrologger.Logger
	Service *cloudformationservice.CloudFormation
}

type Resource struct {
	clients awsclient.Clients
	logger  micrologger.Logger
	service *cloudformationservice.CloudFormation
}

func NewResource(config ResourceConfig) (*Resource, error) {
	if config.Clients.AutoScaling == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Clients.AutoScaling must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}
	if config.Service == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Service must not be empty", config)
	}

	newResource := &Resource{
		clients: config.Clients,
		logger: config.Logger.With(
			"resource", Name,
		),
		service: config.Service,
	}

	return newResource, nil
}

func (r *Resource) Name() string {
	return Name
}
