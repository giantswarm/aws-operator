package cleanuploadbalancers

import (
	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/aws/aws-sdk-go/service/elbv2"
)

type Clients struct {
	ELB   ELBClient
	ELBv2 ELBClientv2
}

// ELBClient describes the methods required to be implemented by an ELB AWS
// client. The ELB API provides support for classic ELBs.
type ELBClient interface {
	DeleteLoadBalancer(*elb.DeleteLoadBalancerInput) (*elb.DeleteLoadBalancerOutput, error)
	DescribeLoadBalancers(*elb.DescribeLoadBalancersInput) (*elb.DescribeLoadBalancersOutput, error)
	DescribeTags(*elb.DescribeTagsInput) (*elb.DescribeTagsOutput, error)
}

// ELBClient describes the methods required to be implemented by an ELB AWS
// client. The ELB API provides support for classic ELBs.
type ELBClientv2 interface {
	DeleteLoadBalancer(*elbv2.DeleteLoadBalancerInput) (*elbv2.DeleteLoadBalancerOutput, error)
	DescribeLoadBalancers(*elbv2.DescribeLoadBalancersInput) (*elbv2.DescribeLoadBalancersOutput, error)
	DescribeTags(*elbv2.DescribeTagsInput) (*elbv2.DescribeTagsOutput, error)
}

type LoadBalancerState struct {
	LoadBalancerNames []string
}

type LoadBalancerStateV2 struct {
	LoadBalancerArns []string
}
