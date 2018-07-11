package loadbalancer

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/aws/aws-sdk-go/service/elb/elbiface"
)

type ELBClientMock struct {
	elbiface.ELBAPI

	loadBalancers []LoadBalancerMock
}

type LoadBalancerMock struct {
	loadBalancerName string
	loadBalancerTags []*elb.Tag
}

func (e *ELBClientMock) DeleteLoadBalancer(*elb.DeleteLoadBalancerInput) (*elb.DeleteLoadBalancerOutput, error) {
	return nil, nil
}

func (e *ELBClientMock) DescribeLoadBalancers(*elb.DescribeLoadBalancersInput) (*elb.DescribeLoadBalancersOutput, error) {
	output := &elb.DescribeLoadBalancersOutput{}
	lbDescs := []*elb.LoadBalancerDescription{}

	for _, lb := range e.loadBalancers {
		lbDesc := &elb.LoadBalancerDescription{
			LoadBalancerName: aws.String(lb.loadBalancerName),
		}
		lbDescs = append(lbDescs, lbDesc)
	}
	output.SetLoadBalancerDescriptions(lbDescs)

	return output, nil
}

func (e *ELBClientMock) DescribeTags(*elb.DescribeTagsInput) (*elb.DescribeTagsOutput, error) {
	output := &elb.DescribeTagsOutput{}
	tagDescs := []*elb.TagDescription{}

	for _, lb := range e.loadBalancers {
		tagDesc := &elb.TagDescription{
			LoadBalancerName: aws.String(lb.loadBalancerName),
			Tags:             lb.loadBalancerTags,
		}
		tagDescs = append(tagDescs, tagDesc)
	}
	output.SetTagDescriptions(tagDescs)

	return output, nil
}
