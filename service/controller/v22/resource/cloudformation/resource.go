package cloudformation

import (
	awscloudformation "github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"

	"github.com/giantswarm/aws-operator/pkg/awstags"
	"github.com/giantswarm/aws-operator/service/controller/v22/adapter"
	"github.com/giantswarm/aws-operator/service/controller/v22/encrypter"
	"github.com/giantswarm/aws-operator/service/controller/v22/key"
)

const (
	// Name is the identifier of the resource.
	Name = "cloudformationv22"
)

type AWSConfig struct {
	AccessKeyID     string
	AccessKeySecret string
	SessionToken    string
	Region          string
	accountID       string
}

// Config represents the configuration used to create a new cloudformation
// resource.
type Config struct {
	APIWhitelist         adapter.APIWhitelist
	HostClients          *adapter.Clients
	Logger               micrologger.Logger
	EncrypterRoleManager encrypter.RoleManager

	AdvancedMonitoringEC2      bool
	EncrypterBackend           string
	GuestPrivateSubnetMaskBits int
	GuestPublicSubnetMaskBits  int
	InstallationName           string
	PublicRouteTables          string
	Route53Enabled             bool
}

// Resource implements the cloudformation resource.
type Resource struct {
	apiWhiteList         adapter.APIWhitelist
	encrypterRoleManager encrypter.RoleManager
	hostClients          *adapter.Clients
	logger               micrologger.Logger

	encrypterBackend           string
	guestPrivateSubnetMaskBits int
	guestPublicSubnetMaskBits  int
	installationName           string
	monitoring                 bool
	publicRouteTables          string
	route53Enabled             bool
}

// New creates a new configured cloudformation resource.
func New(config Config) (*Resource, error) {
	if config.HostClients == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.HostClients must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.Logger must not be empty")
	}
	if config.EncrypterBackend == "" {
		return nil, microerror.Maskf(invalidConfigError, "config.EncrypterBackend must not be empty")
	}
	// GuestPrivateSubnetMaskBits && GuestPublicSubnetMaskBits has been
	// validated on upper level because all IPAM related configuration
	// information is present there.

	newService := &Resource{
		apiWhiteList:         config.APIWhitelist,
		hostClients:          config.HostClients,
		logger:               config.Logger,
		encrypterRoleManager: config.EncrypterRoleManager,

		encrypterBackend:           config.EncrypterBackend,
		guestPrivateSubnetMaskBits: config.GuestPrivateSubnetMaskBits,
		guestPublicSubnetMaskBits:  config.GuestPublicSubnetMaskBits,
		installationName:           config.InstallationName,
		monitoring:                 config.AdvancedMonitoringEC2,
		publicRouteTables:          config.PublicRouteTables,
		route53Enabled:             config.Route53Enabled,
	}

	return newService, nil
}

func (r *Resource) Name() string {
	return Name
}

func (r *Resource) getCloudFormationTags(customObject v1alpha1.AWSConfig) []*awscloudformation.Tag {
	tags := key.ClusterTags(customObject, r.installationName)
	return awstags.NewCloudFormation(tags)
}

func toCreateStackInput(v interface{}) (awscloudformation.CreateStackInput, error) {
	if v == nil {
		return awscloudformation.CreateStackInput{}, nil
	}

	createStackInput, ok := v.(awscloudformation.CreateStackInput)
	if !ok {
		return awscloudformation.CreateStackInput{}, microerror.Maskf(wrongTypeError, "expected '%T', got '%T'", createStackInput, v)
	}

	return createStackInput, nil
}

func toDeleteStackInput(v interface{}) (awscloudformation.DeleteStackInput, error) {
	if v == nil {
		return awscloudformation.DeleteStackInput{}, nil
	}

	deleteStackInput, ok := v.(awscloudformation.DeleteStackInput)
	if !ok {
		return awscloudformation.DeleteStackInput{}, microerror.Maskf(wrongTypeError, "expected '%T', got '%T'", deleteStackInput, v)
	}

	return deleteStackInput, nil
}

func toStackState(v interface{}) (StackState, error) {
	if v == nil {
		return StackState{}, nil
	}

	stackState, ok := v.(StackState)
	if !ok {
		return StackState{}, microerror.Maskf(wrongTypeError, "expected '%T', got '%T'", stackState, v)
	}

	return stackState, nil
}

func toUpdateStackInput(v interface{}) (awscloudformation.UpdateStackInput, error) {
	if v == nil {
		return awscloudformation.UpdateStackInput{}, nil
	}

	updateStackInput, ok := v.(awscloudformation.UpdateStackInput)
	if !ok {
		return awscloudformation.UpdateStackInput{}, microerror.Maskf(wrongTypeError, "expected '%T', got '%T'", updateStackInput, v)
	}

	return updateStackInput, nil
}
