package loadbalancer

import (
	"context"

	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/service/awsconfig/v7/key"
)

const (
	cloudProviderClusterTagValue = "owned"
	cloudProviderServiceTagKey   = "kubernetes.io/service-name"
	loadBalancerTagChunkSize     = 20
)

func (r *Resource) GetCurrentState(ctx context.Context, obj interface{}) (interface{}, error) {
	customObject, err := key.ToCustomObject(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	currentState, err := r.clusterLoadBalancers(customObject)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return currentState, nil
}

func (r *Resource) clusterLoadBalancers(customObject v1alpha1.AWSConfig) (*LoadBalancerState, error) {
	lbState := &LoadBalancerState{}
	clusterLBNames := []string{}

	// We get all load balancers because the API does not allow tag filters.
	output, err := r.clients.ELB.DescribeLoadBalancers(&elb.DescribeLoadBalancersInput{})
	if err != nil {
		return nil, microerror.Mask(err)
	}

	allLBNames := []*string{}
	for _, lb := range output.LoadBalancerDescriptions {
		allLBNames = append(allLBNames, lb.LoadBalancerName)
	}

	// Get loadbalancer tags in batches due to API restriction.
	for i := 0; i < len(allLBNames); i += loadBalancerTagChunkSize {
		endPos := i + loadBalancerTagChunkSize

		if endPos > len(allLBNames) {
			endPos = len(allLBNames)
		}

		lbNames := allLBNames[i:endPos]
		tagsInput := &elb.DescribeTagsInput{
			LoadBalancerNames: lbNames,
		}
		tagsOutput, err := r.clients.ELB.DescribeTags(tagsInput)
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
