package cloudformation

import (
	"github.com/aws/aws-sdk-go/aws"
	awscloudformation "github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"

	"github.com/giantswarm/aws-operator/service/controller/v14patch3/adapter"
	"github.com/giantswarm/aws-operator/service/controller/v14patch3/controllercontext"
	"github.com/giantswarm/aws-operator/service/controller/v14patch3/encrypter"
	"github.com/giantswarm/aws-operator/service/controller/v14patch3/key"
)

const (
	// Name is the identifier of the resource.
	Name = "cloudformationv14patch3"
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

	AdvancedMonitoringEC2 bool
	EncrypterBackend      string
	InstallationName      string
	PublicRouteTables     string
	Route53Enabled        bool
}

// Resource implements the cloudformation resource.
type Resource struct {
	apiWhiteList         adapter.APIWhitelist
	encrypterRoleManager encrypter.RoleManager
	hostClients          *adapter.Clients
	logger               micrologger.Logger

	encrypterBackend  string
	installationName  string
	monitoring        bool
	publicRouteTables string
	route53Enabled    bool
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

	newService := &Resource{
		apiWhiteList:         config.APIWhitelist,
		hostClients:          config.HostClients,
		logger:               config.Logger,
		encrypterRoleManager: config.EncrypterRoleManager,

		encrypterBackend:  config.EncrypterBackend,
		installationName:  config.InstallationName,
		monitoring:        config.AdvancedMonitoringEC2,
		publicRouteTables: config.PublicRouteTables,
		route53Enabled:    config.Route53Enabled,
	}

	return newService, nil
}

func (r *Resource) Name() string {
	return Name
}

func (r *Resource) getCloudFormationTags(customObject v1alpha1.AWSConfig) []*awscloudformation.Tag {
	clusterTags := key.ClusterTags(customObject, r.installationName)
	stackTags := []*awscloudformation.Tag{}

	for k, v := range clusterTags {
		tag := &awscloudformation.Tag{
			Key:   aws.String(k),
			Value: aws.String(v),
		}

		stackTags = append(stackTags, tag)
	}

	return stackTags
}

func (r *Resource) addRoleAccess(sc *controllercontext.Context, customObject v1alpha1.AWSConfig) error {
	accountID, err := sc.AWSService.GetAccountID()
	if err != nil {
		return microerror.Mask(err)
	}

	masterRoleARN := key.MasterRoleARN(customObject, accountID)
	workerRoleARN := key.WorkerRoleARN(customObject, accountID)

	err = r.encrypterRoleManager.AddIAMRoleToAuth(encrypter.DecrypterVaultRole, masterRoleARN, workerRoleARN)
	if err != nil {
		return microerror.Mask(err)
	}
	return nil
}

func (r *Resource) removeRoleAccess(sc *controllercontext.Context, customObject v1alpha1.AWSConfig) error {
	accountID, err := sc.AWSService.GetAccountID()
	if err != nil {
		return microerror.Mask(err)
	}

	masterRoleARN := key.MasterRoleARN(customObject, accountID)
	workerRoleARN := key.WorkerRoleARN(customObject, accountID)

	err = r.encrypterRoleManager.RemoveIAMRoleFromAuth(encrypter.DecrypterVaultRole, masterRoleARN, workerRoleARN)
	if err != nil {
		return microerror.Mask(err)
	}
	return nil
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
