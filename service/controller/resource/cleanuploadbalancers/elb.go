package cleanuploadbalancers

import (
	"context"

	"github.com/aws/aws-sdk-go/service/elb"
	infrastructurev1alpha3 "github.com/giantswarm/apiextensions/v6/pkg/apis/infrastructure/v1alpha3"
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/v2/service/controller/controllercontext"
	"github.com/giantswarm/aws-operator/v2/service/controller/key"
)

func (r *Resource) clusterClassicLoadBalancers(ctx context.Context, customObject infrastructurev1alpha3.AWSCluster) (*LoadBalancerState, error) {
	lbState := &LoadBalancerState{}
	clusterLBNames := []string{}

	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	// We get all load balancers because the API does not allow tag filters.
	output, err := cc.Client.TenantCluster.AWS.ELB.DescribeLoadBalancers(&elb.DescribeLoadBalancersInput{})
	if err != nil {
		return nil, microerror.Mask(err)
	}

	allLBNames := []*string{}
	for _, lb := range output.LoadBalancerDescriptions {
		allLBNames = append(allLBNames, lb.LoadBalancerName)
	}

	lbChunks := splitLoadBalancers(allLBNames, loadBalancerTagChunkSize)

	for _, lbNames := range lbChunks {
		tagsInput := &elb.DescribeTagsInput{
			LoadBalancerNames: lbNames,
		}
		tagsOutput, err := cc.Client.TenantCluster.AWS.ELB.DescribeTags(tagsInput)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		// We filter based on the AWS cloud provider tags to find load balancers
		// associated with the cluster being processed.
		for _, lb := range tagsOutput.TagDescriptions {
			if containsClusterTag(lb.Tags, customObject) && containsServiceTag(lb.Tags) {
				clusterLBNames = append(clusterLBNames, *lb.LoadBalancerName)
			}
		}
	}

	lbState.LoadBalancerNames = clusterLBNames

	return lbState, nil
}

func containsClusterTag(tags []*elb.Tag, customObject infrastructurev1alpha3.AWSCluster) bool {
	tagKey := key.ClusterCloudProviderTag(&customObject)

	for _, tag := range tags {
		if *tag.Key == tagKey && *tag.Value == cloudProviderClusterTagValue {
			return true
		}
	}
	return false
}

func containsServiceTag(tags []*elb.Tag) bool {
	for _, tag := range tags {
		if *tag.Key == cloudProviderServiceTagKey {
			return true
		}
	}

	return false
}
