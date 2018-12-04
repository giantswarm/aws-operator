package adapter

import (
	"fmt"

	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/service/controller/v21/key"
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
	EtcdElbHealthCheckTarget         string
	EtcdElbName                      string
	EtcdElbPortsToOpen               []GuestLoadBalancersAdapterPortPair
	EtcdElbScheme                    string
	EtcdElbSecurityGroupID           string
	ELBHealthCheckHealthyThreshold   int
	ELBHealthCheckInterval           int
	ELBHealthCheckTimeout            int
	ELBHealthCheckUnhealthyThreshold int
	IngressElbHealthCheckTarget      string
	IngressElbName                   string
	IngressElbPortsToOpen            []GuestLoadBalancersAdapterPortPair
	IngressElbScheme                 string
	MasterInstanceResourceName       string
	PublicSubnets                    []string
	PrivateSubnets                   []string
}

func (a *GuestLoadBalancersAdapter) Adapt(cfg Config) error {
	{
		numAZs := len(key.StatusAvailabilityZones(cfg.CustomObject))
		if numAZs < 1 {
			return microerror.Maskf(invalidConfigError, "at least one configured availability zone required")
		}
	}

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

	// etcd load balancer settings.
	etcdElbName, err := key.LoadBalancerName(key.EtcdDomain(cfg.CustomObject), cfg.CustomObject)
	if err != nil {
		return microerror.Mask(err)
	}

	a.EtcdElbHealthCheckTarget = heathCheckTarget(key.EtcdPort(cfg.CustomObject))
	a.EtcdElbName = etcdElbName
	a.EtcdElbPortsToOpen = []GuestLoadBalancersAdapterPortPair{
		{
			PortELB:      key.EtcdPort(cfg.CustomObject),
			PortInstance: key.EtcdPort(cfg.CustomObject),
		},
	}
	a.EtcdElbScheme = internalELBScheme

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
	a.MasterInstanceResourceName = cfg.StackState.MasterInstanceResourceName

	for i := 0; i < len(key.StatusAvailabilityZones(cfg.CustomObject)); i++ {
		a.PublicSubnets = append(a.PublicSubnets, key.PublicSubnetName(i))
		a.PrivateSubnets = append(a.PrivateSubnets, key.PrivateSubnetName(i))
	}

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
