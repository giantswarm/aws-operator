package tccp

import (
	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"

	"github.com/giantswarm/aws-operator/service/controller/internal/changedetection"
	"github.com/giantswarm/aws-operator/service/controller/internal/encrypter"
)

const (
	// Name is the identifier of the resource.
	Name = "tccp"
)

const (
	// namedIAMCapability is the AWS specific capability necessary to work with
	// our Cloud Formation templates. It is required for creating worker policy
	// IAM roles.
	namedIAMCapability = "CAPABILITY_NAMED_IAM"
)

// Config represents the configuration used to create a new cloudformation
// resource.
type Config struct {
	// EncrypterRoleManager manages role encryption. This can be supported by
	// different implementations and thus is optional.
	EncrypterRoleManager encrypter.RoleManager
	G8sClient            versioned.Interface
	Logger               micrologger.Logger

	APIWhitelist       APIWhitelist
	CIDRBlockAWSCNI    string
	Detection          *changedetection.TCCP
	InstallationName   string
	InstanceMonitoring bool
	PublicRouteTables  string
	Route53Enabled     bool
}

// Resource implements the cloudformation resource.
type Resource struct {
	encrypterRoleManager encrypter.RoleManager
	g8sClient            versioned.Interface
	logger               micrologger.Logger

	apiWhiteList       APIWhitelist
	cidrBlockAWSCNI    string
	detection          *changedetection.TCCP
	installationName   string
	instanceMonitoring bool
	publicRouteTables  string
	route53Enabled     bool
}

// New creates a new configured cloudformation resource.
func New(config Config) (*Resource, error) {
	if config.Detection == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Detection must not be empty", config)
	}
	if config.G8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.G8sClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}
	if config.APIWhitelist.Private.Enabled && config.APIWhitelist.Private.SubnetList == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.APIWhitelist.Private.SubnetList must not be empty when %T.APIWhitelist.Private is enabled", config)
	}
	if config.APIWhitelist.Public.Enabled && config.APIWhitelist.Public.SubnetList == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.APIWhitelist.Public.SubnetList must not be empty when %T.APIWhitelist.Public is enabled", config)
	}

	if config.CIDRBlockAWSCNI == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.CIDRBlockAWSCNI must not be empty", config)
	}

	r := &Resource{
		g8sClient:            config.G8sClient,
		detection:            config.Detection,
		encrypterRoleManager: config.EncrypterRoleManager,
		logger:               config.Logger,

		apiWhiteList:       config.APIWhitelist,
		cidrBlockAWSCNI:    config.CIDRBlockAWSCNI,
		installationName:   config.InstallationName,
		instanceMonitoring: config.InstanceMonitoring,
		publicRouteTables:  config.PublicRouteTables,
		route53Enabled:     config.Route53Enabled,
	}

	return r, nil
}

func (r *Resource) Name() string {
	return Name
}
