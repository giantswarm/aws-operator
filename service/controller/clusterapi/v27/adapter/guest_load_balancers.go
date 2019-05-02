package adapter

import (
	"fmt"

	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v27/legacykey"
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
		numAZs := len(legacykey.StatusAvailabilityZones(cfg.CustomObject))
		if numAZs < 1 {
			return microerror.Maskf(invalidConfigError, "at least one configured availability zone required")
		}
	}

	// API load balancer settings.
	apiElbName, err := legacykey.LoadBalancerName(cfg.CustomObject.Spec.Cluster.Kubernetes.API.Domain, cfg.CustomObject)
	if err != nil {
		return microerror.Mask(err)
	}

	a.APIElbHealthCheckTarget = heathCheckTarget(cfg.CustomObject.Spec.Cluster.Kubernetes.API.SecurePort)
	a.APIElbName = apiElbName
	a.APIElbPortsToOpen = []GuestLoadBalancersAdapterPortPair{
		{
			PortELB:      legacykey.KubernetesAPISecurePort(cfg.CustomObject),
			PortInstance: legacykey.KubernetesAPISecurePort(cfg.CustomObject),
		},
	}
	a.APIElbScheme = externalELBScheme

	// etcd load balancer settings.
	etcdElbName, err := legacykey.LoadBalancerName(legacykey.EtcdDomain(cfg.CustomObject), cfg.CustomObject)
	if err != nil {
		return microerror.Mask(err)
	}

	a.EtcdElbHealthCheckTarget = heathCheckTarget(legacykey.EtcdPort(cfg.CustomObject))
	a.EtcdElbName = etcdElbName
	a.EtcdElbPortsToOpen = []GuestLoadBalancersAdapterPortPair{
		{
			PortELB:      legacykey.EtcdPort(cfg.CustomObject),
			PortInstance: legacykey.EtcdPort(cfg.CustomObject),
		},
	}
	a.EtcdElbScheme = internalELBScheme

	// Ingress load balancer settings.
	ingressElbName, err := legacykey.LoadBalancerName(cfg.CustomObject.Spec.Cluster.Kubernetes.IngressController.Domain, cfg.CustomObject)
	if err != nil {
		return microerror.Mask(err)
	}

	a.IngressElbHealthCheckTarget = heathCheckTarget(legacykey.IngressControllerSecurePort)
	a.IngressElbName = ingressElbName
	a.IngressElbPortsToOpen = []GuestLoadBalancersAdapterPortPair{
		{
			PortELB: httpsPort,

			PortInstance: legacykey.IngressControllerSecurePort,
		},
		{
			PortELB:      httpPort,
			PortInstance: legacykey.IngressControllerInsecurePort,
		},
	}
	a.IngressElbScheme = externalELBScheme

	// Load balancer health check settings.
	a.ELBHealthCheckHealthyThreshold = healthCheckHealthyThreshold
	a.ELBHealthCheckInterval = healthCheckInterval
	a.ELBHealthCheckTimeout = healthCheckTimeout
	a.ELBHealthCheckUnhealthyThreshold = healthCheckUnhealthyThreshold
	a.MasterInstanceResourceName = cfg.StackState.MasterInstanceResourceName

	for i := 0; i < len(legacykey.StatusAvailabilityZones(cfg.CustomObject)); i++ {
		a.PublicSubnets = append(a.PublicSubnets, legacykey.PublicSubnetName(i))
		a.PrivateSubnets = append(a.PrivateSubnets, legacykey.PrivateSubnetName(i))
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
