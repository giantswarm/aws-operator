package cloudtags

import (
	"context"
	"strings"

	"github.com/giantswarm/k8sclient/v5/pkg/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apiv1alpha2 "sigs.k8s.io/cluster-api/api/v1alpha2"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/aws-operator/pkg/label"
)

const keyCloudPrefix = "tag.provider.giantswarm.io/"

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

// GetTagsByCluster the cloud tags from CAPI Cluster CR
func (ct *CloudTags) GetTagsByCluster(ctx context.Context, clusterID string) (map[string]string, error) {
	var err error

	tags, err := ct.lookupCloudTags(ctx, clusterID)
	if err != nil {
		return tags, microerror.Mask(err)
	}

	return tags, nil
}

func (ct *CloudTags) lookupCloudTags(ctx context.Context, clusterID string) (map[string]string, error) {
	var list apiv1alpha2.ClusterList
	tags := map[string]string{}

	err := ct.k8sClient.CtrlClient().List(
		ctx,
		&list,
		client.InNamespace(metav1.NamespaceDefault),
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
		if isCloudTagKey(lkey) {
			nkey := trimCloudTagKey(lkey)
			tags[nkey] = lvalue
		}
	}

	return tags, nil
}

// IsCloudTagKey check is a tag with proper prefix
func isCloudTagKey(tagKey string) bool {
	return strings.HasPrefix(tagKey, keyCloudPrefix)
}

// TrimCloudTagKey check is a tag with proper prefix
func trimCloudTagKey(tagKey string) string {
	return strings.Replace(tagKey, keyCloudPrefix, "", 1)
}
