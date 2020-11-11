package tccp

import (
	"github.com/giantswarm/apiextensions/v3/pkg/clientset/versioned"
	"github.com/giantswarm/k8sclient/v5/pkg/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"

	"github.com/giantswarm/aws-operator/service/internal/changedetection"
	"github.com/giantswarm/aws-operator/service/internal/hamaster"
	event "github.com/giantswarm/aws-operator/service/internal/recorder"
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
	Event     event.Interface
	G8sClient versioned.Interface
	HAMaster  hamaster.Interface
	K8sClient k8sclient.Interface
	Logger    micrologger.Logger

	APIWhitelist       ConfigAPIWhitelist
	CIDRBlockAWSCNI    string
	Detection          *changedetection.TCCP
	InstallationName   string
	InstanceMonitoring bool
	PublicRouteTables  string
	Route53Enabled     bool
}

// Resource implements the cloudformation resource.
type Resource struct {
	event     event.Interface
	g8sClient versioned.Interface
	haMaster  hamaster.Interface
	k8sClient k8sclient.Interface
	logger    micrologger.Logger

	apiWhitelist       ConfigAPIWhitelist
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
	if config.Event == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Event must not be empty", config)
	}
	if config.G8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.G8sClient must not be empty", config)
	}
	if config.HAMaster == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.HAMaster must not be empty", config)
	}
	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.K8sClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}
	if config.APIWhitelist.Private.Enabled && len(config.APIWhitelist.Private.SubnetList) == 0 {
		return nil, microerror.Maskf(invalidConfigError, "%T.APIWhitelist.Private.SubnetList must not be empty when %T.APIWhitelist.Private is enabled", config, config)
	}
	if config.APIWhitelist.Public.Enabled && len(config.APIWhitelist.Public.SubnetList) == 0 {
		return nil, microerror.Maskf(invalidConfigError, "%T.APIWhitelist.Public.SubnetList must not be empty when %T.APIWhitelist.Public is enabled", config, config)
	}

	if config.CIDRBlockAWSCNI == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.CIDRBlockAWSCNI must not be empty", config)
	}

	r := &Resource{
		event:     config.Event,
		g8sClient: config.G8sClient,
		haMaster:  config.HAMaster,
		detection: config.Detection,
		k8sClient: config.K8sClient,
		logger:    config.Logger,

		apiWhitelist:       config.APIWhitelist,
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
