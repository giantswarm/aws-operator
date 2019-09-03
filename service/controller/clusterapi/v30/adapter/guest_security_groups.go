package adapter

import (
	"fmt"
	"strings"

	"github.com/giantswarm/microerror"
	"sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"

	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v30/key"
)

const (
	allPorts             = -1
	cadvisorPort         = 4194
	etcdPort             = 2379
	kubeletPort          = 10250
	nodeExporterPort     = 10300
	kubeStateMetricsPort = 10301
	sshPort              = 22

	allProtocols = "-1"
	tcpProtocol  = "tcp"

	defaultCIDR = "0.0.0.0/0"

	ingressSecurityGroupName = "IngressSecurityGroup"
)

type GuestSecurityGroupsAdapter struct {
	APIWhitelistEnabled       bool
	MasterSecurityGroupName   string
	MasterSecurityGroupRules  []securityGroupRule
	IngressSecurityGroupName  string
	IngressSecurityGroupRules []securityGroupRule
	EtcdELBSecurityGroupName  string
	EtcdELBSecurityGroupRules []securityGroupRule
}

func (s *GuestSecurityGroupsAdapter) Adapt(cfg Config) error {
	masterRules, err := s.getMasterRules(cfg, cfg.ControlPlaneVPCCidr)
	if err != nil {
		return microerror.Mask(err)
	}

	s.APIWhitelistEnabled = cfg.APIWhitelist.Enabled

	s.MasterSecurityGroupName = key.SecurityGroupName(&cfg.CustomObject, "master")
	s.MasterSecurityGroupRules = masterRules

	s.IngressSecurityGroupName = key.SecurityGroupName(&cfg.CustomObject, "ingress")
	s.IngressSecurityGroupRules = s.getIngressRules(cfg.CustomObject)

	s.EtcdELBSecurityGroupName = key.SecurityGroupName(&cfg.CustomObject, "etcd-elb")
	s.EtcdELBSecurityGroupRules = s.getEtcdRules(cfg.CustomObject, cfg.ControlPlaneVPCCidr)

	return nil
}

func (s *GuestSecurityGroupsAdapter) getMasterRules(cfg Config, hostClusterCIDR string) ([]securityGroupRule, error) {
	// Allow traffic to the Kubernetes API server depending on the API
	// whitelisting rules.
	apiRules, err := getKubernetesAPIRules(cfg, hostClusterCIDR)
	if err != nil {
		return []securityGroupRule{}, microerror.Mask(err)
	}

	// Other security group rules for the master.
	otherRules := []securityGroupRule{
		{
			Description: "Allow traffic from control plane CIDR to 4194 for cadvisor scraping.",
			Port:        cadvisorPort,
			Protocol:    tcpProtocol,
			SourceCIDR:  hostClusterCIDR,
		},
		{
			Description: "Allow traffic from control plane CIDR to 2379 for etcd backup.",
			Port:        etcdPort,
			Protocol:    tcpProtocol,
			SourceCIDR:  hostClusterCIDR,
		},
		{
			Description: "Allow traffic from control plane CIDR to 10250 for kubelet scraping.",
			Port:        kubeletPort,
			Protocol:    tcpProtocol,
			SourceCIDR:  hostClusterCIDR,
		},
		{
			Description: "Allow traffic from control plane CIDR to 10300 for node-exporter scraping.",
			Port:        nodeExporterPort,
			Protocol:    tcpProtocol,
			SourceCIDR:  hostClusterCIDR,
		},
		{
			Description: "Allow traffic from control plane CIDR to 10301 for kube-state-metrics scraping.",
			Port:        kubeStateMetricsPort,
			Protocol:    tcpProtocol,
			SourceCIDR:  hostClusterCIDR,
		},
		{
			Description: "Only allow ssh traffic from the control plane.",
			Port:        sshPort,
			Protocol:    tcpProtocol,
			SourceCIDR:  hostClusterCIDR,
		},
	}

	return append(apiRules, otherRules...), nil
}

func (s *GuestSecurityGroupsAdapter) getIngressRules(customObject v1alpha1.Cluster) []securityGroupRule {
	return []securityGroupRule{
		{
			Description: "Allow all http traffic to the ingress load balancer.",
			Port:        httpPort,
			Protocol:    tcpProtocol,
			SourceCIDR:  defaultCIDR,
		},
		{
			Description: "Allow all https traffic to the ingress load balancer.",
			Port:        httpsPort,
			Protocol:    tcpProtocol,
			SourceCIDR:  defaultCIDR,
		},
	}
}

func (s *GuestSecurityGroupsAdapter) getEtcdRules(customObject v1alpha1.Cluster, hostClusterCIDR string) []securityGroupRule {
	return []securityGroupRule{
		{
			Description: "Allow all etcd traffic from the VPC to the etcd load balancer.",
			Port:        etcdPort,
			Protocol:    tcpProtocol,
			SourceCIDR:  defaultCIDR,
		},
		{
			Description: "Allow traffic from control plane to etcd port for backup and metrics.",
			Port:        etcdPort,
			Protocol:    tcpProtocol,
			SourceCIDR:  hostClusterCIDR,
		},
	}
}

type securityGroupRule struct {
	Description         string
	Port                int
	Protocol            string
	SourceCIDR          string
	SourceSecurityGroup string
}

func getKubernetesAPIRules(cfg Config, hostClusterCIDR string) ([]securityGroupRule, error) {
	// When API whitelisting is enabled, add separate security group rule per each subnet.
	if cfg.APIWhitelist.Enabled {
		rules := []securityGroupRule{
			{
				Description: "Allow traffic from control plane CIDR.",
				Port:        key.KubernetesSecurePort,
				Protocol:    tcpProtocol,
				SourceCIDR:  hostClusterCIDR,
			},
			{
				Description: "Allow traffic from tenant cluster CIDR.",
				Port:        key.KubernetesSecurePort,
				Protocol:    tcpProtocol,
				SourceCIDR:  key.StatusClusterNetworkCIDR(cfg.CustomObject),
			},
		}

		// Whitelist all configured subnets.
		whitelistSubnets := strings.Split(cfg.APIWhitelist.SubnetList, ",")
		for _, subnet := range whitelistSubnets {
			if subnet != "" {
				subnetRule := securityGroupRule{
					Description: "Custom Whitelist CIDR.",
					Port:        key.KubernetesSecurePort,
					Protocol:    tcpProtocol,
					SourceCIDR:  subnet,
				}
				rules = append(rules, subnetRule)
			}
		}

		// Whitelist public EIPs of the host cluster NAT gateways.
		hostClusterNATGatewayRules, err := getHostClusterNATGatewayRules(cfg)
		if err != nil {
			return []securityGroupRule{}, microerror.Mask(err)
		}

		for _, gatewayRule := range hostClusterNATGatewayRules {
			rules = append(rules, gatewayRule)
		}

		return rules, nil
	} else {
		// When API whitelisting is disabled, allow all traffic.
		allowAllRule := []securityGroupRule{
			{
				Description: "Allow all traffic to the master instance.",
				Port:        key.KubernetesSecurePort,
				Protocol:    tcpProtocol,
				SourceCIDR:  defaultCIDR,
			},
		}

		return allowAllRule, nil
	}
}

func getHostClusterNATGatewayRules(cfg Config) ([]securityGroupRule, error) {
	var gatewayRules []securityGroupRule

	for _, address := range cfg.ControlPlaneNATGatewayAddresses {
		gatewayRule := securityGroupRule{
			Description: "Allow traffic from gateways.",
			Port:        key.KubernetesSecurePort,
			Protocol:    tcpProtocol,
			SourceCIDR:  fmt.Sprintf("%s/32", *address.PublicIp),
		}

		gatewayRules = append(gatewayRules, gatewayRule)
	}

	return gatewayRules, nil
}
