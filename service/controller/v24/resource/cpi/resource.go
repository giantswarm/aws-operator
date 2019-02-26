package cpi

import (
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"

	"github.com/giantswarm/aws-operator/pkg/awstags"
	"github.com/giantswarm/aws-operator/service/controller/v24/adapter"
	"github.com/giantswarm/aws-operator/service/controller/v24/key"
)

const (
	// Name is the identifier of the resource.
	Name = "cpiv24"
)

type Config struct {
	HostClients *adapter.Clients
	Logger      micrologger.Logger

	InstallationName string
	Route53Enabled   bool
}

// Resource implements the CPI resource, which stands for Control Plane
// Initializer. This was formerly known as the host pre stack. We manage a
// dedicated CF stack for the IAM role and VPC Peering setup.
type Resource struct {
	hostClients *adapter.Clients
	logger      micrologger.Logger

	installationName string
	route53Enabled   bool
}

func New(config Config) (*Resource, error) {
	if config.HostClients == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.HostClients must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	r := &Resource{
		hostClients: config.HostClients,
		logger:      config.Logger,

		installationName: config.InstallationName,
		route53Enabled:   config.Route53Enabled,
	}

	return r, nil
}

func (r *Resource) Name() string {
	return Name
}

func (r *Resource) getCloudFormationTags(customObject v1alpha1.AWSConfig) []*cloudformation.Tag {
	tags := key.ClusterTags(customObject, r.installationName)
	return awstags.NewCloudFormation(tags)
}
