package lifecycle

import (
	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"

	awsclient "github.com/giantswarm/aws-operator/client/aws"
	cloudformationservice "github.com/giantswarm/aws-operator/service/controller/v9_patch1/cloudformation"
)

const (
	Name = "lifecyclev9_patch1"
)

type ResourceConfig struct {
	AWS       awsclient.Clients
	G8sClient versioned.Interface
	Logger    micrologger.Logger
	Service   *cloudformationservice.CloudFormation
}

type Resource struct {
	aws       awsclient.Clients
	g8sClient versioned.Interface
	logger    micrologger.Logger
	service   *cloudformationservice.CloudFormation
}

func NewResource(config ResourceConfig) (*Resource, error) {
	if config.AWS.AutoScaling == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.AWS.AutoScaling must not be empty", config)
	}
	if config.AWS.EC2 == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.AWS.EC2 must not be empty", config)
	}
	if config.G8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.G8sClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}
	if config.Service == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Service must not be empty", config)
	}

	newResource := &Resource{
		aws:       config.AWS,
		g8sClient: config.G8sClient,
		logger:    config.Logger,
		service:   config.Service,
	}

	return newResource, nil
}

func (r *Resource) Name() string {
	return Name
}
