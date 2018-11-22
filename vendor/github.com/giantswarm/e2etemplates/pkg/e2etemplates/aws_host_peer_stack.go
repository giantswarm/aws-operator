package e2etemplates

import (
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/e2etemplates/internal/render"
)

type AWSHostPeerStackConfig struct {
	Stack       AWSHostPeerStackConfigStack
	RouteTable0 AWSHostPeerStackConfigRouteTable0
	RouteTable1 AWSHostPeerStackConfigRouteTable1
}

type AWSHostPeerStackConfigStack struct {
	Name string
}

type AWSHostPeerStackConfigRouteTable0 struct {
	Name string
}

type AWSHostPeerStackConfigRouteTable1 struct {
	Name string
}

// NewAWSHostPeerStack renders awsHostPeerStackTemplate.
func NewAWSHostPeerStack(config AWSHostPeerStackConfig) (string, error) {
	if config.Stack.Name == "" {
		return "", microerror.Maskf(invalidConfigError, "%T.Stack.Name must not be empty", config)
	}
	if config.RouteTable0.Name == "" {
		return "", microerror.Maskf(invalidConfigError, "%T.RouteTable0.Name must not be empty", config)
	}
	if config.RouteTable1.Name == "" {
		return "", microerror.Maskf(invalidConfigError, "%T.RouteTable1.Name must not be empty", config)
	}

	template, err := render.Render(awsHostPeerStackTemplate, config)
	if err != nil {
		return "", microerror.Mask(err)
	}

	return template, nil
}
