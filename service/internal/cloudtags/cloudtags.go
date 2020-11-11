package cloudtags

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/giantswarm/k8sclient/v5/pkg/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apiv1alpha2 "sigs.k8s.io/cluster-api/api/v1alpha2"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/aws-operator/pkg/label"
	"github.com/giantswarm/aws-operator/service/controller/controllercontext"
	"github.com/giantswarm/aws-operator/service/controller/key"
	"github.com/giantswarm/aws-operator/service/internal/cloudtags/internal/cache"
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

	awsCache  *cache.AWS
	capiCache *cache.CAPI
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

		awsCache:  cache.NewAWS(),
		capiCache: cache.NewCAPI(),
	}

	return l, nil
}

// CloudTagsNotInSync compares current cluster tags with the stack tags
func (ct *CloudTags) CloudTagsNotInSync(ctx context.Context, crGetter key.LabelsGetter, stackType string) (bool, error) {
	var err error

	ctags, err := ct.GetTagsByCluster(ctx, key.ClusterID(crGetter))
	if err != nil {
		return true, microerror.Mask(err)
	}

	var stags map[string]string
	switch stackType {
	case key.StackTCCP:
		stags, err = ct.GetAWSTagsByTCCP(ctx, crGetter)
	case key.StackTCCPN:
		stags, err = ct.GetAWSTagsByTCCPN(ctx, crGetter)
	case key.StackTCNP:
		stags, err = ct.GetAWSTagsByTCPN(ctx, crGetter)
	default:
		return false, noStackTypeFound
	}
	if err != nil {
		return true, microerror.Mask(err)
	}

	tagsEqual := reflect.DeepEqual(ctags, stags)
	if !tagsEqual {
		// Print changed values and new labels in the cluster CR
		for ck, cv := range ctags {
			if sv, ok := stags[ck]; ok {
				ct.logger.LogCtx(ctx,
					"level", "debug",
					"message", "detected a change in cloud tags",
					"reason", fmt.Sprintf("Existing tag changed from %s to %s", sv, cv),
				)
			} else {
				ct.logger.LogCtx(ctx,
					"level", "debug",
					"message", "detected a change in cloud tags",
					"reason", fmt.Sprintf("New tag %s:%s added to cluster CR", ck, cv),
				)
			}
		}
		// Print removed tags from stack
		for sk, sv := range stags {
			if _, ok := ctags[sk]; !ok {
				ct.logger.LogCtx(ctx,
					"level", "debug",
					"message", "detected a change in cloud tags",
					"reason", fmt.Sprintf("Removed tag %s:%s changed from stack", sk, sv),
				)
			}
		}
		return true, nil
	}

	return false, nil
}

// GetTagsByCluster the cloud tags from CAPI Cluster CR
func (ct *CloudTags) GetTagsByCluster(ctx context.Context, clusterID string) (map[string]string, error) {
	var ok bool
	var err error

	var tags map[string]string
	{
		ck := ct.awsCache.Key(ctx, clusterID)

		if ck == "" {
			tags, err = ct.lookupCloudTags(ctx, clusterID)
			if err != nil {
				return tags, microerror.Mask(err)
			}
		} else {
			tags, ok = ct.awsCache.Get(ctx, ck)
			if !ok {
				tags, err := ct.lookupCloudTags(ctx, clusterID)
				if err != nil {
					return tags, microerror.Mask(err)
				}

				ct.awsCache.Set(ctx, ck, tags)
			}
		}
	}

	return tags, nil
}

// GetAWSTagsByTCCPN the cloud tags from AWS Cloud Formation Stack
func (ct *CloudTags) GetAWSTagsByTCCPN(ctx context.Context, crGetter key.LabelsGetter) (map[string]string, error) {
	var ok bool

	clusterID := key.ClusterID(crGetter)
	var tags map[string]string
	{
		ck := ct.awsCache.Key(ctx, clusterID)

		if ck == "" {
			tags, err := ct.lookupAWStagsForTCCPN(ctx, crGetter)
			if err != nil {
				return tags, microerror.Mask(err)
			}
		} else {
			tags, ok = ct.awsCache.Get(ctx, ck)
			if !ok {
				tags, err := ct.lookupAWStagsForTCCPN(ctx, crGetter)
				if err != nil {
					return tags, microerror.Mask(err)
				}

				ct.awsCache.Set(ctx, ck, tags)
			}
		}
	}

	return tags, nil
}

// GetAWSTagsByTCPN the cloud tags from AWS Cloud Formation Stack
func (ct *CloudTags) GetAWSTagsByTCPN(ctx context.Context, crGetter key.LabelsGetter) (map[string]string, error) {
	var ok bool
	var err error

	clusterID := key.ClusterID(crGetter)
	var tags map[string]string
	{
		ck := ct.awsCache.Key(ctx, clusterID)

		if ck == "" {
			tags, err = ct.lookupAWStagsForTCPN(ctx, crGetter)
			if err != nil {
				return tags, microerror.Mask(err)
			}
		} else {
			tags, ok = ct.awsCache.Get(ctx, ck)
			if !ok {
				tags, err = ct.lookupAWStagsForTCPN(ctx, crGetter)
				if err != nil {
					return tags, microerror.Mask(err)
				}

				ct.awsCache.Set(ctx, ck, tags)
			}
		}
	}

	return tags, nil
}

// GetAWSTagsByTCCP the cloud tags from AWS Cloud Formation Stack
func (ct *CloudTags) GetAWSTagsByTCCP(ctx context.Context, crGetter key.LabelsGetter) (map[string]string, error) {
	var ok bool
	var err error

	clusterID := key.ClusterID(crGetter)
	var tags map[string]string
	{
		ck := ct.awsCache.Key(ctx, clusterID)

		if ck == "" {
			tags, err = ct.lookupAWStagsForTCCP(ctx, crGetter)
			if err != nil {
				return tags, microerror.Mask(err)
			}
		} else {
			tags, ok = ct.awsCache.Get(ctx, ck)
			if !ok {
				tags, err = ct.lookupAWStagsForTCCP(ctx, crGetter)
				if err != nil {
					return tags, microerror.Mask(err)
				}

				ct.awsCache.Set(ctx, ck, tags)
			}
		}
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

func (ct *CloudTags) lookupAWStagsForTCCPN(ctx context.Context, crGetter key.LabelsGetter) (map[string]string, error) {
	stackTags := map[string]string{}
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return stackTags, microerror.Mask(err)
	}

	i := &cloudformation.DescribeStacksInput{
		StackName: aws.String(key.StackNameTCCPN(crGetter)),
	}

	o, err := cc.Client.TenantCluster.AWS.CloudFormation.DescribeStacks(i)
	if err != nil {
		return stackTags, microerror.Mask(err)
	}

	for _, v := range o.Stacks[0].Tags {
		if isStackTagKey(*v.Key) {
			continue
		}
		stackTags[*v.Key] = *v.Value
	}

	return stackTags, nil
}

func (ct *CloudTags) lookupAWStagsForTCCP(ctx context.Context, crGetter key.LabelsGetter) (map[string]string, error) {
	stackTags := map[string]string{}
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return stackTags, microerror.Mask(err)
	}

	i := &cloudformation.DescribeStacksInput{
		StackName: aws.String(key.StackNameTCCP(crGetter)),
	}

	o, err := cc.Client.TenantCluster.AWS.CloudFormation.DescribeStacks(i)
	if err != nil {
		return stackTags, microerror.Mask(err)
	}

	for _, v := range o.Stacks[0].Tags {
		if isStackTagKey(*v.Key) {
			continue
		}
		stackTags[*v.Key] = *v.Value
	}

	return stackTags, nil
}

func (ct *CloudTags) lookupAWStagsForTCPN(ctx context.Context, crGetter key.LabelsGetter) (map[string]string, error) {
	stackTags := map[string]string{}
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return stackTags, microerror.Mask(err)
	}

	i := &cloudformation.DescribeStacksInput{
		StackName: aws.String(key.StackNameTCNP(crGetter)),
	}

	o, err := cc.Client.TenantCluster.AWS.CloudFormation.DescribeStacks(i)
	if err != nil {
		return stackTags, microerror.Mask(err)
	}

	for _, v := range o.Stacks[0].Tags {
		if isStackTagKey(*v.Key) {
			continue
		}
		stackTags[*v.Key] = *v.Value
	}

	return stackTags, nil
}

// IsCloudTagKey check is a tag with proper prefix
func isCloudTagKey(tagKey string) bool {
	return strings.HasPrefix(tagKey, keyCloudPrefix)
}

// IsStackTagKey check is a tag is one of the usuals default keys
func isStackTagKey(tagKey string) bool {
	return strings.HasPrefix(tagKey, keyGiantswarmPrefix) || strings.HasPrefix(tagKey, keyKubernetesPrefix)
}

// TrimCloudTagKey check is a tag with proper prefix
func trimCloudTagKey(tagKey string) string {
	return strings.Replace(tagKey, keyCloudPrefix, "", 1)
}
