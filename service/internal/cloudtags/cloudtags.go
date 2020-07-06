package cloudtags

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/giantswarm/k8sclient/v3/pkg/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	apiv1alpha2 "sigs.k8s.io/cluster-api/api/v1alpha2"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/aws-operator/pkg/label"
)

const keyCloudPrefix = "aws-tag/"
const keyGiantswarmPrefix = "giantswarm.io/"
const keyKubernetesPrefix = "kubernetes.io/"

type Config struct {
	K8sClient k8sclient.Interface
	Logger    micrologger.Logger
}

type CloudTags struct {
	k8sClient k8sclient.Interface
	logger    micrologger.Logger
}

// New CloudTags object
func New(config Config) (*CloudTags, error) {
	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.K8sClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	l := &CloudTags{
		k8sClient: config.K8sClient,
		logger:    config.Logger,
	}

	return l, nil
}

// AreClusterTagsEquals compares current cluster tags with the input
func (ct *CloudTags) AreClusterTagsEquals(ctx context.Context, clusterID string, tags map[string]string) (bool, error) {
	ctags, err := ct.GetTagsByCluster(ctx, clusterID)
	if err != nil {
		return false, microerror.Mask(err)
	}

	tagsEqual := reflect.DeepEqual(ctags, tags)
	if !tagsEqual {
		ct.logger.LogCtx(ctx,
			"level", "debug",
			"message", "detected a change in cloud tags",
			"reason", fmt.Sprintf("tags changed from %#q to %#q", tags, ctags),
		)
		return true, nil
	}

	return false, nil
}

// GetTagsByCluster the cloud tags from CAPI Cluster CR
func (ct *CloudTags) GetTagsByCluster(ctx context.Context, clusterID string) (map[string]string, error) {
	var list apiv1alpha2.ClusterList
	tags := map[string]string{}

	err := ct.k8sClient.CtrlClient().List(
		ctx,
		&list,
		client.InNamespace("default"),
		client.MatchingLabels{label.Cluster: clusterID},
	)
	if err != nil {
		return tags, microerror.Mask(err)
	}

	if len(list.Items) == 0 {
		return tags, microerror.Mask(notFoundError)
	}
	if len(list.Items) > 1 {
		return tags, microerror.Mask(tooManyCRsError)
	}

	labels := list.Items[0].GetLabels()
	for lkey, lvalue := range labels {
		if IsCloudTagKey(lkey) {
			nkey := TrimCloudTagKey(lkey)
			tags[nkey] = lvalue
		}
	}

	return tags, nil
}

// IsCloudTagKey check is a tag with proper prefix
func IsCloudTagKey(tagKey string) bool {
	return strings.HasPrefix(tagKey, keyCloudPrefix)
}

// IsStackTagKey check is a tag is one of the usuals default keys
func IsStackTagKey(tagKey string) bool {
	return strings.HasPrefix(tagKey, keyGiantswarmPrefix) || strings.HasPrefix(tagKey, keyKubernetesPrefix)
}

// TrimCloudTagKey check is a tag with proper prefix
func TrimCloudTagKey(tagKey string) string {
	return strings.Replace(tagKey, keyCloudPrefix, "", 1)
}
