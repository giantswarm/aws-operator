package adapter

import (
	"fmt"
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v31/key"
)

const (
	// Default values for health checks.
	healthCheckHealthyThreshold   = 2
	healthCheckInterval           = 10
	healthCheckTimeout            = 6
	healthCheckUnhealthyThreshold = 2

	// resource names for ELBv2 Listener resources
	apiInternalELBListenerResourceName     = "ApiInternalLoadBalancerListener"
	apiELBListenerResourceName             = "ApiLoadBalancerListener"
	etcdELBListenerResourceName            = "EtcdLoadBalancerListener"
	ingressELBInsecureListenerResourceName = "IngressLoadBalancerInsecureListener"
	ingressELBSecureListenerResourceName   = "IngressLoadBalancerSecureListener"
	// resource names for ELBv2 Target group resources
	apiInternalELBTargetGroupResourceName     = "ApiInternalLoadBalancerTargetGroup"
	apiELBTargetGroupResourceName             = "ApiLoadBalancerTargetGroup"
	etcdELBTargetGroupResourceName            = "EtcdLoadBalancerTargetGroup"
	ingressELBInsecureTargetGroupResourceName = "IngressLoadBalancerInsecureTargetGroup"
	ingressELBSecureTargetGroupResourceName   = "IngressLoadBalancerSecureTargetGroup"
)

type GuestLoadBalancersAdapter struct {
	APIElbHealthCheckTarget           string
	APIElbName                        string
	APIInternalElbName                string
	APIElbListenersAndTargets         []GuestLoadBalancersAdapterListenerAndTarget
	APIInternalElbListenersAndTargets []GuestLoadBalancersAdapterListenerAndTarget
	APIElbScheme                      string
	APIInternalElbScheme              string
	APIElbSecurityGroupID             string
	EtcdElbHealthCheckTarget          string
	EtcdElbName                       string
	EtcdElbListenersAndTargets        []GuestLoadBalancersAdapterListenerAndTarget
	EtcdElbScheme                     string
	EtcdElbSecurityGroupID            string
	ELBHealthCheckHealthyThreshold    int
	ELBHealthCheckInterval            int
	ELBHealthCheckTimeout             int
	ELBHealthCheckUnhealthyThreshold  int
	IngressElbHealthCheckTarget       string
	IngressElbName                    string
	IngressElbListenersAndTargets     []GuestLoadBalancersAdapterListenerAndTarget
	IngressElbScheme                  string
	MasterInstanceResourceName        string
	PublicSubnets                     []string
	PrivateSubnets                    []string
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
	a.APIElbListenersAndTargets = []GuestLoadBalancersAdapterListenerAndTarget{
		{
			ListenerResourceName: apiELBListenerResourceName,
			PortELB:              key.KubernetesSecurePort,
			PortInstance:         key.KubernetesSecurePort,
			TargetResourceName:   apiELBTargetGroupResourceName,
		},
	}
	a.APIInternalElbListenersAndTargets = []GuestLoadBalancersAdapterListenerAndTarget{
		{
			ListenerResourceName: apiInternalELBListenerResourceName,
			PortELB:              key.KubernetesSecurePort,
			PortInstance:         key.KubernetesSecurePort,
			TargetResourceName:   apiInternalELBTargetGroupResourceName,
		},
	}
	a.APIElbScheme = externalELBScheme
	a.APIInternalElbScheme = internalELBScheme

	// etcd load balancer settings.
	a.EtcdElbHealthCheckTarget = heathCheckTarget(key.EtcdPort)
	a.EtcdElbName = key.ELBNameEtcd(&cfg.CustomObject)
	a.EtcdElbListenersAndTargets = []GuestLoadBalancersAdapterListenerAndTarget{
		{
			ListenerResourceName: etcdELBListenerResourceName,
			PortELB:              key.EtcdPort,
			PortInstance:         key.EtcdPort,
			TargetResourceName:   etcdELBTargetGroupResourceName,
		},
	}
	a.EtcdElbScheme = internalELBScheme

	// Ingress load balancer settings.
	a.IngressElbHealthCheckTarget = heathCheckTarget(key.IngressControllerSecurePort)
	a.IngressElbName = key.ELBNameIngress(&cfg.CustomObject)
	a.IngressElbListenersAndTargets = []GuestLoadBalancersAdapterListenerAndTarget{
		{
			ListenerResourceName: ingressELBSecureListenerResourceName,
			PortELB:              httpsPort,
			PortInstance:         key.IngressControllerSecurePort,
			TargetResourceName:   ingressELBSecureTargetGroupResourceName,
		},
		{
			ListenerResourceName: ingressELBInsecureListenerResourceName,
			PortELB:              httpPort,
			PortInstance:         key.IngressControllerInsecurePort,
			TargetResourceName:   ingressELBInsecureTargetGroupResourceName,
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

type GuestLoadBalancersAdapterListenerAndTarget struct {
	// Listener resource name
	ListenerResourceName string
	// PortELB is the port the ELB should listen on.
	PortELB int
	// PortInstance is the target port on the instance the ELB forwards traffic to.
	PortInstance int
	// Target Group resource name
	TargetResourceName string
}

func heathCheckTarget(port int) string {
	return fmt.Sprintf("TCP:%d", port)
}
