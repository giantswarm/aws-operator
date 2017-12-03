package s3bucketv1

import (
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/framework"
)

const (
	// Name is the identifier of the resource.
	Name = "s3bucket"
)

// Config represents the configuration used to create a new s3bucket resource.
type Config struct {
	// Dependencies.
	Logger   micrologger.Logger
	S3Client *s3.S3
}

// DefaultConfig provides a default configuration to create a new s3bucket
// resource by best effort.
func DefaultConfig() Config {
	return Config{
		// Dependencies.
		Logger:   nil,
		S3Client: nil,
	}
}

// Resource implements the s3bucket resource.
type Resource struct {
	// Dependencies.
	s3Client *s3.S3
	logger   micrologger.Logger
}

// New creates a new configured s3bucket resource.
func New(config Config) (*Resource, error) {
	// Dependencies.
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.Logger must not be empty")
	}
	if config.S3Client == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.S3Client must not be empty")
	}

	newResource := &Resource{
		// Dependencies.
		logger: config.Logger.With(
			"resource", Name,
		),
		s3Client: config.S3Client,
	}

	return newResource, nil
}

func (r *Resource) Name() string {
	return Name
}

func (r *Resource) Underlying() framework.Resource {
	return r
}
