package tccp

import (
	"github.com/giantswarm/k8sclient/v7/pkg/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"

	"github.com/giantswarm/aws-operator/service/internal/changedetection"
	"github.com/giantswarm/aws-operator/service/internal/cloudtags"
	"github.com/giantswarm/aws-operator/service/internal/hamaster"
	event "github.com/giantswarm/aws-operator/service/internal/recorder"

	ctrlClient "sigs.k8s.io/controller-runtime/pkg/client"
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
	CloudTags  cloudtags.Interface
	Event      event.Interface
	CtrlClient ctrlClient.Client
	HAMaster   hamaster.Interface
	K8sClient  k8sclient.Interface
	Logger     micrologger.Logger

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
	cloudtags  cloudtags.Interface
	event      event.Interface
	ctrlClient ctrlClient.Client
	haMaster   hamaster.Interface
	k8sClient  k8sclient.Interface
	logger     micrologger.Logger

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
	if config.CloudTags == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.CloudTags must not be empty", config)
	}
	if config.Detection == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Detection must not be empty", config)
	}
	if config.Event == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Event must not be empty", config)
	}
	if config.CtrlClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.CtrlClient must not be empty", config)
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
		cloudtags:  config.CloudTags,
		event:      config.Event,
		ctrlClient: config.CtrlClient,
		haMaster:   config.HAMaster,
		detection:  config.Detection,
		k8sClient:  config.K8sClient,
		logger:     config.Logger,

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
