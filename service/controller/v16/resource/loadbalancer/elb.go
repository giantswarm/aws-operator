package loadbalancer

import (
	"context"

	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/service/controller/v16/controllercontext"
	"github.com/giantswarm/aws-operator/service/controller/v16/key"
)

const (
	cloudProviderClusterTagValue = "owned"
	cloudProviderServiceTagKey   = "kubernetes.io/service-name"
	loadBalancerTagChunkSize     = 20
)

func (r *Resource) clusterLoadBalancers(ctx context.Context, customObject v1alpha1.AWSConfig) (*LoadBalancerState, error) {
	lbState := &LoadBalancerState{}
	clusterLBNames := []string{}

	sc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	// We get all load balancers because the API does not allow tag filters.
	output, err := sc.AWSClient.ELB.DescribeLoadBalancers(&elb.DescribeLoadBalancersInput{})
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
		tagsOutput, err := sc.AWSClient.ELB.DescribeTags(tagsInput)
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

func splitLoadBalancers(loadBalancerNames []*string, chunkSize int) [][]*string {
	chunks := make([][]*string, 0)

	for i := 0; i < len(loadBalancerNames); i += chunkSize {
		endPos := i + chunkSize

		if endPos > len(loadBalancerNames) {
			endPos = len(loadBalancerNames)
		}

		chunks = append(chunks, loadBalancerNames[i:endPos])
	}

	return chunks
}

func containsClusterTag(tags []*elb.Tag, customObject v1alpha1.AWSConfig) bool {
	tagKey := key.ClusterCloudProviderTag(customObject)

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
