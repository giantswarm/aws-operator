package adapter

import (
	"fmt"

	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/service/controller/v14patch1/key"
)

const (
	// Default values for health checks.
	healthCheckHealthyThreshold   = 2
	healthCheckInterval           = 5
	healthCheckTimeout            = 3
	healthCheckUnhealthyThreshold = 2
)

type GuestLoadBalancersAdapter struct {
	APIElbHealthCheckTarget          string
	APIElbName                       string
	APIElbPortsToOpen                []GuestLoadBalancersAdapterPortPair
	APIElbScheme                     string
	APIElbSecurityGroupID            string
	ELBHealthCheckHealthyThreshold   int
	ELBHealthCheckInterval           int
	ELBHealthCheckTimeout            int
	ELBHealthCheckUnhealthyThreshold int
	IngressElbHealthCheckTarget      string
	IngressElbName                   string
	IngressElbPortsToOpen            []GuestLoadBalancersAdapterPortPair
	IngressElbScheme                 string
	MasterInstanceResourceName       string
}

func (a *GuestLoadBalancersAdapter) Adapt(cfg Config) error {
	// API load balancer settings.
	apiElbName, err := key.LoadBalancerName(cfg.CustomObject.Spec.Cluster.Kubernetes.API.Domain, cfg.CustomObject)
	if err != nil {
		return microerror.Mask(err)
	}

	a.APIElbHealthCheckTarget = heathCheckTarget(cfg.CustomObject.Spec.Cluster.Kubernetes.API.SecurePort)
	a.APIElbName = apiElbName
	a.APIElbPortsToOpen = []GuestLoadBalancersAdapterPortPair{
		{
			PortELB:      key.KubernetesAPISecurePort(cfg.CustomObject),
			PortInstance: key.KubernetesAPISecurePort(cfg.CustomObject),
		},
	}
	a.APIElbScheme = externalELBScheme

	// Ingress load balancer settings.
	ingressElbName, err := key.LoadBalancerName(cfg.CustomObject.Spec.Cluster.Kubernetes.IngressController.Domain, cfg.CustomObject)
	if err != nil {
		return microerror.Mask(err)
	}

	a.IngressElbHealthCheckTarget = heathCheckTarget(key.IngressControllerSecurePort(cfg.CustomObject))
	a.IngressElbName = ingressElbName
	a.IngressElbPortsToOpen = []GuestLoadBalancersAdapterPortPair{
		{
			PortELB: httpsPort,

			PortInstance: key.IngressControllerSecurePort(cfg.CustomObject),
		},
		{
			PortELB:      httpPort,
			PortInstance: key.IngressControllerInsecurePort(cfg.CustomObject),
		},
	}
	a.IngressElbScheme = externalELBScheme

	// Load balancer health check settings.
	a.ELBHealthCheckHealthyThreshold = healthCheckHealthyThreshold
	a.ELBHealthCheckInterval = healthCheckInterval
	a.ELBHealthCheckTimeout = healthCheckTimeout
	a.ELBHealthCheckUnhealthyThreshold = healthCheckUnhealthyThreshold
	a.MasterInstanceResourceName = masterInstanceResourceName(cfg)

	return nil
}

type GuestLoadBalancersAdapterPortPair struct {
	// PortELB is the port the ELB should listen on.
	PortELB int
	// PortInstance is the port on the instance the ELB forwards traffic to.
	PortInstance int
}

func heathCheckTarget(port int) string {
	return fmt.Sprintf("TCP:%d", port)
}
