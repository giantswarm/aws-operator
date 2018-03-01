package lifecycle

import (
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"

	awsclient "github.com/giantswarm/aws-operator/client/aws"
)

const (
	Name = "lifecyclev7"
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
	if config.Clients.AutoScaling == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Clients.AutoScaling must not be empty", config)
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

func getStackOutputValue(outputs []*cloudformation.Output, key string) (string, error) {
	for _, o := range outputs {
		if *o.OutputKey == key {
			return *o.OutputValue, nil
		}
	}

	return "", microerror.Maskf(notFoundError, "stack outpout value for key '%s'", key)
}
