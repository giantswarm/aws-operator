package adapter

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/service/keyv2"
)

const (
	// Default values for health checks.
	healthCheckHealthyThreshold   = 10
	healthCheckInterval           = 5
	healthCheckTimeout            = 3
	healthCheckUnhealthyThreshold = 2

	httpPort  = 80
	httpsPort = 443
)

type loadBalancersAdapter struct {
	APIElbHealthCheckTarget          string
	APIElbIdleTimoutSeconds          int
	APIElbName                       string
	APIElbPortsToOpen                portPairs
	APIElbScheme                     string
	APIElbSecurityGroupID            string
	ELBHealthCheckHealthyThreshold   int
	ELBHealthCheckInterval           int
	ELBHealthCheckTimeout            int
	ELBHealthCheckUnhealthyThreshold int
	IngressElbHealthCheckTarget      string
	IngressElbIdleTimoutSeconds      int
	IngressElbName                   string
	IngressElbPortsToOpen            portPairs
	IngressElbScheme                 string
	IngressElbSecurityGroupID        string
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

func (lb *loadBalancersAdapter) getLoadBalancers(cfg Config) error {
	// API load balancer settings.
	apiElbName, err := keyv2.LoadBalancerName(cfg.CustomObject.Spec.Cluster.Kubernetes.API.Domain, cfg.CustomObject)
	if err != nil {
		return microerror.Mask(err)
	}

	lb.APIElbHealthCheckTarget = heathCheckTarget(cfg.CustomObject.Spec.Cluster.Kubernetes.API.SecurePort)
	lb.APIElbIdleTimoutSeconds = cfg.CustomObject.Spec.AWS.API.ELB.IdleTimeoutSeconds
	lb.APIElbName = apiElbName
	lb.APIElbPortsToOpen = portPairs{
		{
			PortELB:      cfg.CustomObject.Spec.Cluster.Kubernetes.API.SecurePort,
			PortInstance: cfg.CustomObject.Spec.Cluster.Kubernetes.API.SecurePort,
		},
	}
	lb.APIElbScheme = externalELBScheme

	// Ingress load balancer settings.
	ingressElbName, err := keyv2.LoadBalancerName(cfg.CustomObject.Spec.Cluster.Kubernetes.IngressController.Domain, cfg.CustomObject)
	if err != nil {
		return microerror.Mask(err)
	}

	lb.IngressElbHealthCheckTarget = heathCheckTarget(cfg.CustomObject.Spec.Cluster.Kubernetes.IngressController.SecurePort)
	lb.IngressElbIdleTimoutSeconds = cfg.CustomObject.Spec.AWS.Ingress.ELB.IdleTimeoutSeconds
	lb.IngressElbName = ingressElbName
	lb.IngressElbPortsToOpen = portPairs{
		{
			PortELB:      httpsPort,
			PortInstance: cfg.CustomObject.Spec.Cluster.Kubernetes.IngressController.SecurePort,
		},
		{
			PortELB:      httpPort,
			PortInstance: cfg.CustomObject.Spec.Cluster.Kubernetes.IngressController.InsecurePort,
		},
	}
	lb.IngressElbScheme = externalELBScheme

	// Load balancer health check settings.
	lb.ELBHealthCheckHealthyThreshold = healthCheckHealthyThreshold
	lb.ELBHealthCheckInterval = healthCheckInterval
	lb.ELBHealthCheckTimeout = healthCheckTimeout
	lb.ELBHealthCheckUnhealthyThreshold = healthCheckUnhealthyThreshold

	// master security group field.
	// TODO: remove this code once the security group is created by cloudformation
	// and add a reference in the template
	masterGroupName := keyv2.SecurityGroupName(cfg.CustomObject, prefixMaster)
	describeSgInput := &ec2.DescribeSecurityGroupsInput{
		Filters: []*ec2.Filter{
			{
				Name: aws.String(subnetDescription),
				Values: []*string{
					aws.String(masterGroupName),
				},
			},
			{
				Name: aws.String(subnetGroupName),
				Values: []*string{
					aws.String(masterGroupName),
				},
			},
		},
	}
	output, err := cfg.Clients.EC2.DescribeSecurityGroups(describeSgInput)
	if err != nil {
		return microerror.Mask(err)
	}
	if len(output.SecurityGroups) > 1 {
		return microerror.Mask(tooManyResultsError)
	}
	lb.APIElbSecurityGroupID = *output.SecurityGroups[0].GroupId

	// ingress security group field.
	// TODO: remove this code once the security group is created by cloudformation
	// and add a reference in the template
	ingressGroupName := keyv2.SecurityGroupName(cfg.CustomObject, prefixIngress)
	describeSgInput = &ec2.DescribeSecurityGroupsInput{
		Filters: []*ec2.Filter{
			{
				Name: aws.String(subnetDescription),
				Values: []*string{
					aws.String(ingressGroupName),
				},
			},
			{
				Name: aws.String(subnetGroupName),
				Values: []*string{
					aws.String(ingressGroupName),
				},
			},
		},
	}
	outputIngress, err := cfg.Clients.EC2.DescribeSecurityGroups(describeSgInput)
	if err != nil {
		return microerror.Mask(err)
	}
	if len(outputIngress.SecurityGroups) > 1 {
		return microerror.Mask(tooManyResultsError)
	}
	lb.IngressElbSecurityGroupID = *outputIngress.SecurityGroups[0].GroupId

	return nil
}

func heathCheckTarget(port int) string {
	return fmt.Sprintf("TCP:%d", port)
}
