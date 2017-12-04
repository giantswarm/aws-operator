package s3bucketv1

import (
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/framework"

	awsutil "github.com/giantswarm/aws-operator/client/aws"
	awsservice "github.com/giantswarm/aws-operator/service/aws"
)

const (
	// Name is the identifier of the resource.
	Name = "s3bucket"
)

// Config represents the configuration used to create a new s3bucket resource.
type Config struct {
	// Dependencies.
	AwsService *awsservice.Service
	Clients    Clients
	Logger     micrologger.Logger

	// Settings.
	AwsConfig awsutil.Config
}

// DefaultConfig provides a default configuration to create a new s3bucket
// resource by best effort.
func DefaultConfig() Config {
	return Config{
		// Dependencies.
		AwsService: nil,
		Clients:    Clients{},
		Logger:     nil,

		// Settings.
		AwsConfig: awsutil.Config{},
	}
}

// Resource implements the s3bucket resource.
type Resource struct {
	// Dependencies.
	awsService *awsservice.Service
	clients    Clients
	logger     micrologger.Logger

	// Settings.
	awsConfig awsutil.Config
}

// New creates a new configured s3bucket resource.
func New(config Config) (*Resource, error) {
	// Dependencies.
	if config.AwsService == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.AwsService must not be empty")
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.Logger must not be empty")
	}

	// Settings.
	var emptyAwsConfig awsutil.Config
	if config.AwsConfig == emptyAwsConfig {
		return nil, microerror.Maskf(invalidConfigError, "config.AwsConfig must not be empty")
	}

	newResource := &Resource{
		// Dependencies.
		awsService: config.AwsService,
		clients:    config.Clients,
		logger: config.Logger.With(
			"resource", Name,
		),

		// Settings.
		awsConfig: config.AwsConfig,
	}

	return newResource, nil
}

func (r *Resource) Name() string {
	return Name
}

func (r *Resource) Underlying() framework.Resource {
	return r
}
