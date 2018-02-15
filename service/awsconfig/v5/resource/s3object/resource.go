package s3object

import (
	"context"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/giantswarm/certs/legacy"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/framework"
)

const (
	// Name is the identifier of the resource.
	Name = "s3objectv5"
)

// Config represents the configuration used to create a new cloudformation resource.
type Config struct {
	// Dependencies.
	AwsService       AwsService
	CertWatcher      legacy.Searcher
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
	certWatcher      legacy.Searcher
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
	return framework.NewPatch(), nil
}

func toBucketObjectState(v interface{}) (map[string]BucketObjectState, error) {
	if v == nil {
		return map[string]BucketObjectState{}, nil
	}

	bucketObjectState, ok := v.(map[string]BucketObjectState)
	if !ok {
		return map[string]BucketObjectState{}, microerror.Maskf(wrongTypeError, "expected '%T', got '%T'", bucketObjectState, v)
	}

	return bucketObjectState, nil
}

func toPutObjectInput(v interface{}) (s3.PutObjectInput, error) {
	if v == nil {
		return s3.PutObjectInput{}, nil
	}

	bucketObject, ok := v.(BucketObjectState)
	if !ok {
		return s3.PutObjectInput{}, microerror.Maskf(wrongTypeError, "expected '%T', got '%T'", bucketObject, v)
	}

	putObjectInput := s3.PutObjectInput{
		Key:           aws.String(bucketObject.Key),
		Body:          strings.NewReader(bucketObject.Body),
		Bucket:        aws.String(bucketObject.Bucket),
		ContentLength: aws.Int64(int64(len(bucketObject.Body))),
	}

	return putObjectInput, nil
}
