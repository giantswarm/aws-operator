package s3bucket

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"

	awsservice "github.com/giantswarm/aws-operator/service/aws"
	"github.com/giantswarm/aws-operator/service/awsconfig/v5/key"
)

const (
	// Name is the identifier of the resource.
	Name = "s3bucketv5"
)

// Config represents the configuration used to create a new s3bucket resource.
type Config struct {
	// Dependencies.
	AwsService *awsservice.Service
	Clients    Clients
	Logger     micrologger.Logger
}

// DefaultConfig provides a default configuration to create a new s3bucket
// resource by best effort.
func DefaultConfig() Config {
	return Config{
		// Dependencies.
		AwsService: nil,
		Clients:    Clients{},
		Logger:     nil,
	}
}

// Resource implements the s3bucket resource.
type Resource struct {
	// Dependencies.
	awsService *awsservice.Service
	clients    Clients
	logger     micrologger.Logger
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

	newResource := &Resource{
		// Dependencies.
		awsService: config.AwsService,
		clients:    config.Clients,
		logger: config.Logger.With(
			"resource", Name,
		),
	}

	return newResource, nil
}

func (r *Resource) Name() string {
	return Name
}

func toBucketState(v interface{}) (BucketState, error) {
	if v == nil {
		return BucketState{}, nil
	}

	bucketState, ok := v.(BucketState)
	if !ok {
		return BucketState{}, microerror.Maskf(wrongTypeError, "expected '%T', got '%T'", BucketState{}, v)
	}

	return bucketState, nil
}

func getS3BucketTags(customObject v1alpha1.AWSConfig) []*s3.Tag {
	clusterTags := key.ClusterTags(customObject)
	s3Tags := []*s3.Tag{}

	for k, v := range clusterTags {
		tag := &s3.Tag{
			Key:   aws.String(k),
			Value: aws.String(v),
		}

		s3Tags = append(s3Tags, tag)
	}

	return s3Tags
}
