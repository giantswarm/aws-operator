package adapter

import (
	"fmt"

	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/service/controller/v4/key"
)

// The template related to this adapter can be found in the following import.
//
//     github.com/giantswarm/aws-operator/service/controller/v4/templates/cloudformation/guest/load_balancers.go
//

const (
	// Default values for health checks.
	healthCheckHealthyThreshold   = 2
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

func (lb *loadBalancersAdapter) getLoadBalancers(cfg Config) error {
	// API load balancer settings.
	apiElbName, err := key.LoadBalancerName(cfg.CustomObject.Spec.Cluster.Kubernetes.API.Domain, cfg.CustomObject)
	if err != nil {
		return microerror.Mask(err)
	}

	lb.APIElbHealthCheckTarget = heathCheckTarget(cfg.CustomObject.Spec.Cluster.Kubernetes.API.SecurePort)
	lb.APIElbIdleTimoutSeconds = cfg.CustomObject.Spec.AWS.API.ELB.IdleTimeoutSeconds
	lb.APIElbName = apiElbName
	lb.APIElbPortsToOpen = portPairs{
		{
			PortELB:      key.KubernetesAPISecurePort(cfg.CustomObject),
			PortInstance: key.KubernetesAPISecurePort(cfg.CustomObject),
		},
	}
	lb.APIElbScheme = externalELBScheme

	// Ingress load balancer settings.
	ingressElbName, err := key.LoadBalancerName(cfg.CustomObject.Spec.Cluster.Kubernetes.IngressController.Domain, cfg.CustomObject)
	if err != nil {
		return microerror.Mask(err)
	}

	lb.IngressElbHealthCheckTarget = heathCheckTarget(key.IngressControllerSecurePort(cfg.CustomObject))
	lb.IngressElbIdleTimoutSeconds = cfg.CustomObject.Spec.AWS.Ingress.ELB.IdleTimeoutSeconds
	lb.IngressElbName = ingressElbName
	lb.IngressElbPortsToOpen = portPairs{
		{
			PortELB: httpsPort,

			PortInstance: key.IngressControllerSecurePort(cfg.CustomObject),
		},
		{
			PortELB:      httpPort,
			PortInstance: key.IngressControllerInsecurePort(cfg.CustomObject),
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
