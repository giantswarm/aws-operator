package loadbalancer

import (
	"github.com/aws/aws-sdk-go/service/elb"
)

type Clients struct {
	ELB ELBClient
}

// ELBClient describes the methods required to be implemented by an ELB AWS
// client. The ELB API provides support for classic ELBs.
type ELBClient interface {
	DeleteLoadBalancer(*elb.DeleteLoadBalancerInput) (*elb.DeleteLoadBalancerOutput, error)
	DescribeLoadBalancers(*elb.DescribeLoadBalancersInput) (*elb.DescribeLoadBalancersOutput, error)
	DescribeTags(*elb.DescribeTagsInput) (*elb.DescribeTagsOutput, error)
}

type LoadBalancerState struct {
	LoadBalancerNames []string
}
