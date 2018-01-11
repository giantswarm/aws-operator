package cloudformationv2

import (
	"github.com/aws/aws-sdk-go/aws"
	awscloudformation "github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/framework"

	"github.com/giantswarm/aws-operator/service/keyv2"
	"github.com/giantswarm/aws-operator/service/resource/cloudformationv2/adapter"
)

const (
	// Name is the identifier of the resource.
	Name = "cloudformationv2"
)

type AWSConfig struct {
	AccessKeyID     string
	AccessKeySecret string
	SessionToken    string
	Region          string
	accountID       string
}

// Config represents the configuration used to create a new cloudformation resource.
type Config struct {
	// Dependencies.
	Clients          *adapter.Clients
	HostClients      *adapter.Clients
	InstallationName string
	Logger           micrologger.Logger
}

// DefaultConfig provides a default configuration to create a new cloudformation
// resource by best effort.
func DefaultConfig() Config {
	return Config{
		// Dependencies.
		Clients:          &adapter.Clients{},
		HostClients:      &adapter.Clients{},
		InstallationName: "",
		Logger:           nil,
	}
}

// Resource implements the cloudformation resource.
type Resource struct {
	// Dependencies.
	Clients          *adapter.Clients
	HostClients      *adapter.Clients
	installationName string
	logger           micrologger.Logger
}

// New creates a new configured cloudformation resource.
func New(config Config) (*Resource, error) {
	// Dependencies.
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.Logger must not be empty")
	}

	newService := &Resource{
		// Dependencies.
		Clients:     config.Clients,
		HostClients: config.HostClients,
		logger: config.Logger.With(
			"resource", Name,
		),
	}

	return newService, nil
}

func (r *Resource) Name() string {
	return Name
}

func (r *Resource) Underlying() framework.Resource {
	return r
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

func getStackOutputValue(outputs []*awscloudformation.Output, key string) (string, error) {
	for _, o := range outputs {
		if *o.OutputKey == key {
			return *o.OutputValue, nil
		}
	}

	return "", microerror.Mask(notFoundError)
}

func (r *Resource) createHostPreStack(customObject v1alpha1.AWSConfig) error {
	stackName := keyv2.MainHostPreStackName(customObject)
	mainTemplate, err := r.getMainHostPreTemplateBody(customObject)
	if err != nil {
		return microerror.Mask(err)
	}
	createStack := &awscloudformation.CreateStackInput{
		StackName:    aws.String(stackName),
		TemplateBody: aws.String(mainTemplate),
		// CAPABILITY_NAMED_IAM is required for creating IAM roles (worker policy)
		Capabilities: []*string{
			aws.String("CAPABILITY_NAMED_IAM"),
		},
	}

	r.logger.Log("debug", "creating AWS Host Pre-Guest cloudformation stack")
	_, err = r.HostClients.CloudFormation.CreateStack(createStack)
	if err != nil {
		return microerror.Mask(err)
	}

	err = r.HostClients.CloudFormation.WaitUntilStackCreateComplete(&awscloudformation.DescribeStacksInput{
		StackName: aws.String(stackName),
	})
	if err != nil {
		return microerror.Mask(err)
	}
	r.logger.Log("debug", "creating AWS Host Pre-Guest cloudformation stack: created")
	return nil
}

func (r *Resource) createHostPostStack(customObject v1alpha1.AWSConfig) error {
	stackName := keyv2.MainHostPostStackName(customObject)
	mainTemplate, err := r.getMainHostPostTemplateBody(customObject)
	if err != nil {
		return microerror.Mask(err)
	}
	createStack := &awscloudformation.CreateStackInput{
		StackName:    aws.String(stackName),
		TemplateBody: aws.String(mainTemplate),
	}

	r.logger.Log("debug", "creating AWS Host Post-Guest cloudformation stack")
	_, err = r.HostClients.CloudFormation.CreateStack(createStack)
	if err != nil {
		return microerror.Mask(err)
	}

	err = r.HostClients.CloudFormation.WaitUntilStackCreateComplete(&awscloudformation.DescribeStacksInput{
		StackName: aws.String(stackName),
	})
	if err != nil {
		return microerror.Mask(err)
	}

	r.logger.Log("debug", "creating AWS Host Post-Guest cloudformation stack: created")

	return nil
}
