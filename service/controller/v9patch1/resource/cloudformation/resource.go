package cloudformation

import (
	"github.com/aws/aws-sdk-go/aws"
	awscloudformation "github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"

	"github.com/giantswarm/aws-operator/service/controller/v9patch1/adapter"
	cloudformationservice "github.com/giantswarm/aws-operator/service/controller/v9patch1/cloudformation"
	"github.com/giantswarm/aws-operator/service/controller/v9patch1/ebs"
	"github.com/giantswarm/aws-operator/service/controller/v9patch1/key"
)

const (
	// Name is the identifier of the resource.
	Name = "cloudformationv9patch1"
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
	APIWhitelist adapter.APIWhitelist
	Clients      *adapter.Clients
	EBS          ebs.Interface
	HostClients  *adapter.Clients
	Logger       micrologger.Logger
	Service      *cloudformationservice.CloudFormation

	InstallationName string
}

// Resource implements the cloudformation resource.
type Resource struct {
	apiWhiteList adapter.APIWhitelist
	clients      *adapter.Clients
	ebs          ebs.Interface
	hostClients  *adapter.Clients
	logger       micrologger.Logger
	service      *cloudformationservice.CloudFormation

	installationName string
}

// New creates a new configured cloudformation resource.
func New(config Config) (*Resource, error) {
	if config.Clients == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Clients must not be empty", config)
	}
	if config.EBS == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.EBS must not be empty", config)
	}
	if config.HostClients == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.HostClients must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.Logger must not be empty")
	}
	if config.Service == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Service must not be empty", config)
	}

	newService := &Resource{
		apiWhiteList: config.APIWhitelist,
		clients:      config.Clients,
		ebs:          config.EBS,
		hostClients:  config.HostClients,
		logger:       config.Logger,
		service:      config.Service,

		installationName: config.InstallationName,
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
