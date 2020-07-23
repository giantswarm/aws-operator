package s3bucket

import (
	"github.com/aws/aws-sdk-go/service/s3"
	infrastructurev1alpha2 "github.com/giantswarm/apiextensions/pkg/apis/infrastructure/v1alpha2"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"

	"github.com/giantswarm/aws-operator/pkg/awstags"
	"github.com/giantswarm/aws-operator/service/controller/key"
	event "github.com/giantswarm/aws-operator/service/internal/recorder"
)

const (
	// Name is the identifier of the resource.
	Name = "s3bucket"
	// LifecycleLoggingBucketID is the Lifecycle ID for the logging bucket
	LifecycleLoggingBucketID = "ExpirationLogs"
)

// Config represents the configuration used to create a new s3bucket resource.
type Config struct {
	// Dependencies.
	Event  event.Interface
	Logger micrologger.Logger

	// Settings.
	AccessLogsExpiration int
	DeleteLoggingBucket  bool
	IncludeTags          bool
	InstallationName     string
}

// Resource implements the s3bucket resource.
type Resource struct {
	// Dependencies.
	event  event.Interface
	logger micrologger.Logger

	// Settings.
	accessLogsExpiration int
	deleteLoggingBucket  bool
	includeTags          bool
	installationName     string
}

// New creates a new configured s3bucket resource.
func New(config Config) (*Resource, error) {
	// Dependencies.
	if config.Event == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Event must not be empty", config)
	}
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

	r := &Resource{
		// Dependencies.
		event:  config.Event,
		logger: config.Logger,

		// Settings.
		accessLogsExpiration: config.AccessLogsExpiration,
		deleteLoggingBucket:  config.DeleteLoggingBucket,
		includeTags:          config.IncludeTags,
		installationName:     config.InstallationName,
	}

	return r, nil
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

func (r *Resource) getS3BucketTags(customObject infrastructurev1alpha2.AWSCluster) []*s3.Tag {
	tags := key.AWSTags(&customObject, r.installationName)
	return awstags.NewS3(tags)
}

func (r *Resource) canBeDeleted(bucket BucketState) bool {
	return !bucket.IsLoggingBucket || bucket.IsLoggingBucket && r.deleteLoggingBucket
}
