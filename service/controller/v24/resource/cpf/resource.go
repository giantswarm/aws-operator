package cpf

import (
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"

	"github.com/giantswarm/aws-operator/pkg/awstags"
	"github.com/giantswarm/aws-operator/service/controller/v24/key"
	"github.com/giantswarm/aws-operator/service/routetable"
)

const (
	// Name is the identifier of the resource.
	Name = "cpfv24"
)

type Config struct {
	CloudFormation CloudFormation
	Logger         micrologger.Logger
	RouteTable     *routetable.RouteTable

	EncrypterBackend  string
	InstallationName  string
	PublicRouteTables string
	Route53Enabled    bool
}

// Resource implements the CPF resource, which stands for Control Plane
// Finalizer. This was formerly known as the host post stack. We manage a
// dedicated CF stack for the record sets and routing tables setup.
type Resource struct {
	cloudFormation CloudFormation
	logger         micrologger.Logger
	routeTable     *routetable.RouteTable

	encrypterBackend  string
	installationName  string
	publicRouteTables string
	route53Enabled    bool
}

func New(config Config) (*Resource, error) {
	if config.CloudFormation == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.CloudFormation must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}
	if config.RouteTable == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.RouteTable must not be empty", config)
	}

	if config.EncrypterBackend == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.EncrypterBackend must not be empty", config)
	}

	r := &Resource{
		cloudFormation: config.CloudFormation,
		logger:         config.Logger,
		routeTable:     config.RouteTable,

		encrypterBackend:  config.EncrypterBackend,
		installationName:  config.InstallationName,
		publicRouteTables: config.PublicRouteTables,
		route53Enabled:    config.Route53Enabled,
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
