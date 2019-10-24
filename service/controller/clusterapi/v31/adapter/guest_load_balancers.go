package adapter

import (
	"fmt"

	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v31/key"
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
	APIInternalElbName               string
	APIElbPortsToOpen                []GuestLoadBalancersAdapterPortPair
	APIElbScheme                     string
	APIInternalElbScheme             string
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
	clusterAZs := cfg.TenantClusterAvailabilityZones
	if len(clusterAZs) < 1 {
		return microerror.Maskf(invalidConfigError, "at least one configured availability zone required")
	}

	// API load balancer settings.
	a.APIElbHealthCheckTarget = heathCheckTarget(key.KubernetesSecurePort)
	a.APIElbName = key.ELBNameAPI(&cfg.CustomObject)
	a.APIInternalElbName = key.InternalELBNameAPI(&cfg.CustomObject)
	a.APIElbPortsToOpen = []GuestLoadBalancersAdapterPortPair{
		{
			PortELB:      key.KubernetesSecurePort,
			PortInstance: key.KubernetesSecurePort,
		},
	}
	a.APIElbScheme = externalELBScheme
	a.APIInternalElbScheme = internalELBScheme

	// etcd load balancer settings.
	a.EtcdElbHealthCheckTarget = heathCheckTarget(key.EtcdPort)
	a.EtcdElbName = key.ELBNameEtcd(&cfg.CustomObject)
	a.EtcdElbPortsToOpen = []GuestLoadBalancersAdapterPortPair{
		{
			PortELB:      key.EtcdPort,
			PortInstance: key.EtcdPort,
		},
	}
	a.EtcdElbScheme = internalELBScheme

	// Ingress load balancer settings.
	a.IngressElbHealthCheckTarget = heathCheckTarget(key.IngressControllerSecurePort)
	a.IngressElbName = key.ELBNameIngress(&cfg.CustomObject)
	a.IngressElbPortsToOpen = []GuestLoadBalancersAdapterPortPair{
		{
			PortELB: httpsPort,

			PortInstance: key.IngressControllerSecurePort,
		},
		{
			PortELB:      httpPort,
			PortInstance: key.IngressControllerInsecurePort,
		},
	}
	a.IngressElbScheme = externalELBScheme

	// Load balancer health check settings.
	a.ELBHealthCheckHealthyThreshold = healthCheckHealthyThreshold
	a.ELBHealthCheckInterval = healthCheckInterval
	a.ELBHealthCheckTimeout = healthCheckTimeout
	a.ELBHealthCheckUnhealthyThreshold = healthCheckUnhealthyThreshold
	a.MasterInstanceResourceName = cfg.StackState.MasterInstanceResourceName

	for _, az := range clusterAZs {
		a.PublicSubnets = append(a.PublicSubnets, key.SanitizeCFResourceName(key.PublicSubnetName(az.Name)))
	}

	for _, az := range clusterAZs {
		if az.Name != key.MasterAvailabilityZone(cfg.CustomObject) {
			continue
		}

		a.PrivateSubnets = append(a.PrivateSubnets, key.SanitizeCFResourceName(key.PrivateSubnetName(az.Name)))
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
