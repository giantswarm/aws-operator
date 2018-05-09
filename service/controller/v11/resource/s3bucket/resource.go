package s3bucket

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"

	"github.com/giantswarm/aws-operator/service/controller/v11/key"
)

const (
	// Name is the identifier of the resource.
	Name = "s3bucketv11"
	// LifecycleLoggingBucketID is the Lifecycle ID for the logging bucket
	LifecycleLoggingBucketID = "ExpirationLogs"
)

// Config represents the configuration used to create a new s3bucket resource.
type Config struct {
	// Dependencies.
	Logger micrologger.Logger

	// Settings.
	AccessLogsExpiration int
	InstallationName     string
}

// DefaultConfig provides a default configuration to create a new s3bucket
// resource by best effort.
func DefaultConfig() Config {
	return Config{
		// Dependencies.
		Logger: nil,
	}
}

// Resource implements the s3bucket resource.
type Resource struct {
	// Dependencies.
	logger micrologger.Logger

	// Settings.
	accessLogsExpiration int
	installationName     string
}

// New creates a new configured s3bucket resource.
func New(config Config) (*Resource, error) {
	// Dependencies.
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	// Settings.
	if config.AccessLogsExpiration < 0 {
		return nil, microerror.Maskf(invalidConfigError, "%T.AccessLogsExpiration must not be lower than 0", config)
	}
	if config.InstallationName == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.InstallationName must not be empty", config)
	}

	newResource := &Resource{
		// Dependencies.
		logger: config.Logger,

		// Settings.
		accessLogsExpiration: config.AccessLogsExpiration,
		installationName:     config.InstallationName,
	}

	return newResource, nil
}

func (r *Resource) Name() string {
	return Name
}

func toBucketState(v interface{}) ([]BucketState, error) {
	if v == nil {
		return []BucketState{}, nil
	}

	bucketsState, ok := v.([]BucketState)
	if !ok {
		return nil, microerror.Maskf(wrongTypeError, "expected '%T', got '%T'", []BucketState{}, v)
	}

	return bucketsState, nil
}

func containsBucketState(bucketStateName string, bucketStateList []BucketState) bool {
	for _, b := range bucketStateList {
		if b.Name == bucketStateName {
			return true
		}
	}

	return false
}

func (r *Resource) getS3BucketTags(customObject v1alpha1.AWSConfig) []*s3.Tag {
	clusterTags := key.ClusterTags(customObject, r.installationName)
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
