package cleanuploadbalancers

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elbv2"
	infrastructurev1alpha3 "github.com/giantswarm/apiextensions/v3/pkg/apis/infrastructure/v1alpha3"
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/service/controller/controllercontext"
	"github.com/giantswarm/aws-operator/service/controller/key"
)

func (r *Resource) clusterLoadBalancersV2(ctx context.Context, cl infrastructurev1alpha3.AWSCluster) (*LoadBalancerStateV2, error) {
	lbState := &LoadBalancerStateV2{}
	clusterLBArns := []string{}

	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	// We get all load balancers because the API does not allow tag filters.
	output, err := cc.Client.TenantCluster.AWS.ELBv2.DescribeLoadBalancers(&elbv2.DescribeLoadBalancersInput{})
	if err != nil {
		return nil, microerror.Mask(err)
	}

	allLBArns := []*string{}
	for _, lb := range output.LoadBalancers {
		allLBArns = append(allLBArns, lb.LoadBalancerArn)
	}

	lbChunks := splitLoadBalancers(allLBArns, loadBalancerTagChunkSize)

	for _, lbArn := range lbChunks {
		tagsInput := &elbv2.DescribeTagsInput{
			ResourceArns: lbArn,
		}
		tagsOutput, err := cc.Client.TenantCluster.AWS.ELBv2.DescribeTags(tagsInput)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		// We filter based on the AWS cloud provider tags to find load balancers
		// associated with the cluster being processed.
		for _, lb := range tagsOutput.TagDescriptions {
			if containsClusterTagV2(lb.Tags, cl) && containsServiceTagV2(lb.Tags) {
				clusterLBArns = append(clusterLBArns, *lb.ResourceArn)
			}
		}
	}

	lbState.LoadBalancerArns = clusterLBArns

	targetGroupsArns := []string{}

	// fetch all related target groups which needs to be deleted as well
	for _, lbArn := range lbState.LoadBalancerArns {
		i := &elbv2.DescribeTargetGroupsInput{
			LoadBalancerArn: aws.String(lbArn),
		}

		o, err := cc.Client.TenantCluster.AWS.ELBv2.DescribeTargetGroups(i)
		if err != nil {
			return nil, microerror.Mask(err)
		}
		for _, targetGroup := range o.TargetGroups {
			targetGroupsArns = append(targetGroupsArns, *targetGroup.TargetGroupArn)
		}

	}
	lbState.TargetGroupsArns = targetGroupsArns

	return lbState, nil
}

func containsClusterTagV2(tags []*elbv2.Tag, customObject infrastructurev1alpha3.AWSCluster) bool {
	tagKey := key.ClusterCloudProviderTag(&customObject)

	for _, tag := range tags {
		if *tag.Key == tagKey && *tag.Value == cloudProviderClusterTagValue {
			return true
		}
	}
	return false
}

func containsServiceTagV2(tags []*elbv2.Tag) bool {
	for _, tag := range tags {
		if *tag.Key == cloudProviderServiceTagKey {
			return true
		}
	}

	return false
}
