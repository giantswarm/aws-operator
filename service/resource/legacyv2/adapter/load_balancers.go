package adapter

import (
	"fmt"

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
	ELBHealthCheckHealthyThreshold   int
	ELBHealthCheckInterval           int
	ELBHealthCheckTimeout            int
	ELBHealthCheckUnhealthyThreshold int
	IngressElbHealthCheckTarget      string
	IngressElbIdleTimoutSeconds      int
	IngressElbName                   string
	IngressElbPortsToOpen            portPairs
	IngressElbScheme                 string
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

	// Ingress load balancer settings.
	ingressElbName, err := keyv2.LoadBalancerName(customObject.Spec.Cluster.Kubernetes.IngressController.Domain, customObject)
	if err != nil {
		return microerror.Mask(err)
	}

	lb.IngressElbHealthCheckTarget = heathCheckTarget(customObject.Spec.Cluster.Kubernetes.IngressController.SecurePort)
	lb.IngressElbIdleTimoutSeconds = customObject.Spec.AWS.Ingress.ELB.IdleTimeoutSeconds
	lb.IngressElbName = ingressElbName
	lb.IngressElbPortsToOpen = portPairs{
		{
			PortELB:      httpsPort,
			PortInstance: keyv2.IngressControllerSecurePort(customObject),
		},
		{
			PortELB:      httpPort,
			PortInstance: keyv2.IngressControllerInsecurePort(customObject),
		},
	}
	lb.IngressElbScheme = externalELBScheme

	// Load balancer health check settings.
	lb.ELBHealthCheckHealthyThreshold = healthCheckHealthyThreshold
	lb.ELBHealthCheckInterval = healthCheckInterval
	lb.ELBHealthCheckTimeout = healthCheckTimeout
	lb.ELBHealthCheckUnhealthyThreshold = healthCheckUnhealthyThreshold

	return nil
}

func heathCheckTarget(port int) string {
	return fmt.Sprintf("TCP:%d", port)
}
