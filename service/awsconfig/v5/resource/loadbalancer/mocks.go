package loadbalancer

import (
	"github.com/aws/aws-sdk-go/service/elb"
)

type ELBClientMock struct {
}

func (e *ELBClientMock) DeleteLoadBalancer(*elb.DeleteLoadBalancerInput) (*elb.DeleteLoadBalancerOutput, error) {
	return nil, nil
}

func (e *ELBClientMock) DescribeLoadBalancers(*elb.DescribeLoadBalancersInput) (*elb.DescribeLoadBalancersOutput, error) {
	return nil, nil
}

func (e *ELBClientMock) DescribeTags(*elb.DescribeTagsInput) (*elb.DescribeTagsOutput, error) {
	return nil, nil
}
