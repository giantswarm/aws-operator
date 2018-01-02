package adapter

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/service/keyv2"
)

const (
	// Default values for health checks.
	healthCheckHealthyThreshold   = 10
	healthCheckInterval           = 5
	healthCheckTimeout            = 3
	healthCheckUnhealthyThreshold = 2
)

type loadBalancersAdapter struct {
	APIElbHealthCheckTarget          string
	APIElbIdleTimoutSeconds          int
	APIElbName                       string
	APIElbPortsToOpen                portPairs
	APIElbScheme                     string
	APIElbSecurityGroupID            string
	APIElbSubnetID                   string
	ELBHealthCheckHealthyThreshold   int
	ELBHealthCheckInterval           int
	ELBHealthCheckTimeout            int
	ELBHealthCheckUnhealthyThreshold int
}

// portPair is a pair of ports.
type portPair struct {
	// PortELB is the port the ELB should listen on.
	PortELB int
	// PortInstance is the port on the instance the ELB forwards traffic to.
	PortInstance int
}

// portPairs is an array of PortPair.
type portPairs []portPair

func (lb *loadBalancersAdapter) getLoadBalancers(customObject v1alpha1.AWSConfig, clients Clients) error {
	// API load balancer settings.
	apiElbName, err := keyv2.LoadBalancerName(customObject.Spec.Cluster.Kubernetes.API.Domain, customObject)
	if err != nil {
		return microerror.Mask(err)
	}

	lb.APIElbHealthCheckTarget = heathCheckTarget(customObject.Spec.Cluster.Kubernetes.API.SecurePort)
	lb.APIElbIdleTimoutSeconds = customObject.Spec.AWS.API.ELB.IdleTimeoutSeconds
	lb.APIElbName = apiElbName
	lb.APIElbPortsToOpen = portPairs{
		{
			PortELB:      customObject.Spec.Cluster.Kubernetes.API.SecurePort,
			PortInstance: customObject.Spec.Cluster.Kubernetes.API.SecurePort,
		},
	}
	lb.APIElbScheme = externalELBScheme

	// Load balancer health check settings.
	lb.ELBHealthCheckHealthyThreshold = healthCheckHealthyThreshold
	lb.ELBHealthCheckInterval = healthCheckInterval
	lb.ELBHealthCheckTimeout = healthCheckTimeout
	lb.ELBHealthCheckUnhealthyThreshold = healthCheckUnhealthyThreshold

	// security group field.
	// TODO: remove this code once the security group is created by cloudformation
	// and add a reference in the template
	groupName := keyv2.SecurityGroupName(customObject, prefixMaster)
	describeSgInput := &ec2.DescribeSecurityGroupsInput{
		Filters: []*ec2.Filter{
			{
				Name: aws.String(subnetDescription),
				Values: []*string{
					aws.String(groupName),
				},
			},
			{
				Name: aws.String(subnetGroupName),
				Values: []*string{
					aws.String(groupName),
				},
			},
		},
	}
	output, err := clients.EC2.DescribeSecurityGroups(describeSgInput)
	if err != nil {
		return microerror.Mask(err)
	}
	if len(output.SecurityGroups) > 1 {
		return microerror.Mask(tooManyResultsError)
	}
	lb.APIElbSecurityGroupID = *output.SecurityGroups[0].GroupId

	// subnet ID
	// TODO: remove this code once the subnet is created by cloudformation and add a
	// reference in the template
	subnetName := keyv2.SubnetName(customObject, suffixPublic)
	describeSubnetInput := &ec2.DescribeSubnetsInput{
		Filters: []*ec2.Filter{
			{
				Name: aws.String(fmt.Sprintf("tag:%s", tagKeyName)),
				Values: []*string{
					aws.String(subnetName),
				},
			},
		},
	}
	outputSubnet, err := clients.EC2.DescribeSubnets(describeSubnetInput)
	if err != nil {
		return microerror.Mask(err)
	}
	if len(outputSubnet.Subnets) > 1 {
		return microerror.Mask(tooManyResultsError)
	}

	lb.APIElbSubnetID = *outputSubnet.Subnets[0].SubnetId

	return nil
}

func heathCheckTarget(port int) string {
	return fmt.Sprintf("TCP:%d", port)
}
