package s3objectv2

import (
	"context"

	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/framework"
)

const (
	// Name is the identifier of the resource.
	Name = "s3objectv2"
)

// Config represents the configuration used to create a new cloudformation resource.
type Config struct {
	// Dependencies.
	AwsService       AwsService
	CertWatcher      CertWatcher
	Clients          Clients
	CloudConfig      CloudConfigService
	Logger           micrologger.Logger
	RandomKeyWatcher RandomKeyWatcher
}

// DefaultConfig provides a default configuration to create a new cloudformation
// resource by best effort.
func DefaultConfig() Config {
	return Config{
		// Dependencies.
		AwsService:       nil,
		CertWatcher:      nil,
		Clients:          Clients{},
		CloudConfig:      nil,
		Logger:           nil,
		RandomKeyWatcher: nil,
	}
}

// Resource implements the cloudformation resource.
type Resource struct {
	// Dependencies.
	awsService       AwsService
	awsClients       Clients
	certWatcher      CertWatcher
	cloudConfig      CloudConfigService
	logger           micrologger.Logger
	randomKeyWatcher RandomKeyWatcher
}

// New creates a new configured cloudformation resource.
func New(config Config) (*Resource, error) {
	// Dependencies.
	if config.AwsService == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.AwsService must not be empty")
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.Logger must not be empty")
	}
	if config.CloudConfig == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.CloudConfig must not be empty")
	}
	if config.CertWatcher == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.CertWatcher must not be empty")
	}
	if config.RandomKeyWatcher == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.RandomKeyWatcher must not be empty")
	}

	newService := &Resource{
		// Dependencies.
		awsService:  config.AwsService,
		awsClients:  config.Clients,
		certWatcher: config.CertWatcher,
		cloudConfig: config.CloudConfig,
		logger: config.Logger.With(
			"resource", Name,
		),
		randomKeyWatcher: config.RandomKeyWatcher,
	}

	return newService, nil
}

func (r *Resource) Name() string {
	return Name
}

func (r *Resource) Underlying() framework.Resource {
	return r
}

func (r *Resource) ApplyDeleteChange(ctx context.Context, obj, deleteChange interface{}) error {
	return nil
}

func (r *Resource) NewDeletePatch(ctx context.Context, obj, currentState, desiredState interface{}) (*framework.Patch, error) {
	return &framework.Patch{}, nil
}

func toBucketObjectState(v interface{}) (BucketObjectState, error) {
	if v == nil {
		return BucketObjectState{}, nil
	}

	bucketObject, ok := v.(BucketObjectState)
	if !ok {
		return BucketObjectState{}, microerror.Maskf(wrongTypeError, "expected '%T', got '%T'", bucketObject, v)
	}

	return bucketObject, nil

}

func toPutObjectInput(v interface{}) (s3.PutObjectInput, error) {
	if v == nil {
		return s3.PutObjectInput{}, nil
	}

	putObjectInput, ok := v.(s3.PutObjectInput)
	if !ok {
		return s3.PutObjectInput{}, microerror.Maskf(wrongTypeError, "expected '%T', got '%T'", putObjectInput, v)
	}

	return putObjectInput, nil
}

func toDeleteObjectInput(v interface{}) (s3.DeleteObjectInput, error) {
	if v == nil {
		return s3.DeleteObjectInput{}, nil
	}

	deleteObjectInput, ok := v.(s3.DeleteObjectInput)
	if !ok {
		return s3.DeleteObjectInput{}, microerror.Maskf(wrongTypeError, "expected '%T', got '%T'", deleteObjectInput, v)
	}

	return deleteObjectInput, nil
}
