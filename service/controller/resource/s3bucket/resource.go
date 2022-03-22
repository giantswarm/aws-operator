package s3bucket

import (
	"context"
	"strings"

	"github.com/aws/aws-sdk-go/service/s3"
	infrastructurev1alpha3 "github.com/giantswarm/apiextensions/v5/pkg/apis/infrastructure/v1alpha3"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	apiv1alpha3 "sigs.k8s.io/cluster-api/api/v1alpha3"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/aws-operator/pkg/awstags"
	"github.com/giantswarm/aws-operator/pkg/label"
	"github.com/giantswarm/aws-operator/service/controller/key"
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
	Logger     micrologger.Logger
	CtrlClient client.Client

	// Settings.
	AccessLogsExpiration int
	DeleteLoggingBucket  bool
	IncludeTags          bool
	InstallationName     string
}

// Resource implements the s3bucket resource.
type Resource struct {
	// Dependencies.
	ctrlClient client.Client
	logger     micrologger.Logger

	// Settings.
	accessLogsExpiration int
	deleteLoggingBucket  bool
	includeTags          bool
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

	r := &Resource{
		// Dependencies.
		ctrlClient: config.CtrlClient,
		logger:     config.Logger,

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

func (r *Resource) getS3BucketTags(ctx context.Context, customObject infrastructurev1alpha3.AWSCluster) ([]*s3.Tag, error) {
	tags := key.AWSTags(&customObject, r.installationName)

	var list apiv1alpha3.ClusterList
	err := r.ctrlClient.List(
		ctx,
		&list,
		client.MatchingLabels{label.Cluster: key.ClusterID(&customObject)},
	)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	if len(list.Items) != 1 {
		return nil, microerror.Maskf(executionFailedError, "expected 1 CR got %d", len(list.Items))
	}

	allTags := addCloudTags(tags, list.Items[0].GetLabels())

	return awstags.NewS3(allTags), nil
}

func (r *Resource) canBeDeleted(bucket BucketState) bool {
	return !bucket.IsLoggingBucket || bucket.IsLoggingBucket && r.deleteLoggingBucket
}

// add cloud tags from `labels` to `tags` map
func addCloudTags(tags map[string]string, labels map[string]string) map[string]string {
	for k, v := range labels {
		if isCloudTagKey(k) {
			tags[trimCloudTagKey(k)] = v
		}
	}

	return tags
}

// IsCloudTagKey checks if a tag has proper prefix
func isCloudTagKey(tagKey string) bool {
	return strings.HasPrefix(tagKey, key.KeyCloudPrefix)
}

// TrimCloudTagKey trims key cloud prefix from a tag
func trimCloudTagKey(tagKey string) string {
	return strings.TrimPrefix(tagKey, key.KeyCloudPrefix)
}
