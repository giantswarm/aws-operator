package cloudformation

import (
	awsCF "github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/cloudformation/cloudformationiface"
	awsutil "github.com/giantswarm/aws-operator/client/aws"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/framework"
)

const (
	// Name is the identifier of the resource.
	Name = "cloudformation"
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
	Clients awsutil.Clients
	Logger  micrologger.Logger
}

// DefaultConfig provides a default configuration to create a new cloudformation
// resource by best effort.
func DefaultConfig() Config {
	return Config{
		// Dependencies.
		Clients: awsutil.Clients{},
		Logger:  nil,
	}
}

// Resource implements the cloudformation resource.
type Resource struct {
	// Dependencies.
	awsClient cloudformationiface.CloudFormationAPI
	logger    micrologger.Logger
}

// StackState is the state representation pn which the resource methods work
type StackState struct {
	Name         string
	TemplateBody string
	Parameters   []*awsCF.Parameter
}

// New creates a new configured cloudformation resource.
func New(config Config) (*Resource, error) {
	// Dependencies.
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.Logger must not be empty")
	}

	newService := &Resource{
		// Dependencies.
		awsClient: config.Clients.CloudFormation,
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
