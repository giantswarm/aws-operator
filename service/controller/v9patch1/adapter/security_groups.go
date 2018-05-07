package adapter

import (
	"strings"

	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/service/controller/v9patch1/key"
)

// The template related to this adapter can be found in the following import.
//
//     github.com/giantswarm/aws-operator/service/controller/v9patch1/templates/cloudformation/guest/security_groups.go
//

type securityGroupsAdapter struct {
	APIWhitelistEnabled       bool
	MasterSecurityGroupName   string
	MasterSecurityGroupRules  []securityGroupRule
	WorkerSecurityGroupName   string
	WorkerSecurityGroupRules  []securityGroupRule
	IngressSecurityGroupName  string
	IngressSecurityGroupRules []securityGroupRule
}

type securityGroupRule struct {
	Port                int
	Protocol            string
	SourceCIDR          string
	SourceSecurityGroup string
}

const (
	allPorts             = -1
	cadvisorPort         = 4194
	kubeletPort          = 10250
	nodeExporterPort     = 10300
	kubeStateMetricsPort = 10301
	sshPort              = 22

	allProtocols = "-1"
	tcpProtocol  = "tcp"

	defaultCIDR = "0.0.0.0/0"

	ingressSecurityGroupName = "IngressSecurityGroup"
)

func (s *securityGroupsAdapter) getSecurityGroups(cfg Config) error {
	hostClusterCIDR, err := VpcCIDR(cfg.HostClients, key.PeerID(cfg.CustomObject))
	if err != nil {
		return microerror.Mask(err)
	}

	s.MasterSecurityGroupName = key.SecurityGroupName(cfg.CustomObject, prefixMaster)
	s.MasterSecurityGroupRules = s.getMasterRules(cfg, hostClusterCIDR)

	s.WorkerSecurityGroupName = key.SecurityGroupName(cfg.CustomObject, prefixWorker)
	s.WorkerSecurityGroupRules = s.getWorkerRules(cfg.CustomObject, hostClusterCIDR)

	s.IngressSecurityGroupName = key.SecurityGroupName(cfg.CustomObject, prefixIngress)
	s.IngressSecurityGroupRules = s.getIngressRules(cfg.CustomObject)

	return nil
}

func (s *securityGroupsAdapter) getMasterRules(cfg Config, hostClusterCIDR string) []securityGroupRule {
	// Allow all traffic to the kubernetes api server.
	apiRules := getKubernetesAPIRule(cfg, hostClusterCIDR)
	// other security group rules for master
	otherRules := []securityGroupRule{
		// Allow traffic from host cluster CIDR to 4194 for cadvisor scraping.
		{
			Port:       cadvisorPort,
			Protocol:   tcpProtocol,
			SourceCIDR: hostClusterCIDR,
		},
		// Allow traffic from host cluster CIDR to 10250 for kubelet scraping.
		{
			Port:       kubeletPort,
			Protocol:   tcpProtocol,
			SourceCIDR: hostClusterCIDR,
		},
		// Allow traffic from host cluster CIDR to 10300 for node-exporter scraping.
		{
			Port:       nodeExporterPort,
			Protocol:   tcpProtocol,
			SourceCIDR: hostClusterCIDR,
		},
		// Allow traffic from host cluster CIDR to 10301 for kube-state-metrics scraping.
		{
			Port:       kubeStateMetricsPort,
			Protocol:   tcpProtocol,
			SourceCIDR: hostClusterCIDR,
		},
		// Only allow ssh traffic from the host cluster.
		{
			Port:       sshPort,
			Protocol:   tcpProtocol,
			SourceCIDR: hostClusterCIDR,
		},
	}
	return append(apiRules, otherRules...)
}

func (s *securityGroupsAdapter) getWorkerRules(customObject v1alpha1.AWSConfig, hostClusterCIDR string) []securityGroupRule {
	return []securityGroupRule{
		// Allow traffic from the ingress security group to the ingress controller.
		{
			Port:                key.IngressControllerSecurePort(customObject),
			Protocol:            tcpProtocol,
			SourceSecurityGroup: ingressSecurityGroupName,
		},
		{
			Port:                key.IngressControllerInsecurePort(customObject),
			Protocol:            tcpProtocol,
			SourceSecurityGroup: ingressSecurityGroupName,
		},
		// Allow traffic from host cluster to ingress controller secure port,
		// for guest cluster scraping.
		{
			Port:       key.IngressControllerSecurePort(customObject),
			Protocol:   tcpProtocol,
			SourceCIDR: hostClusterCIDR,
		},
		// Allow traffic from host cluster CIDR to 4194 for cadvisor scraping.
		{
			Port:       cadvisorPort,
			Protocol:   tcpProtocol,
			SourceCIDR: hostClusterCIDR,
		},
		// Allow traffic from host cluster CIDR to 10250 for kubelet scraping.
		{
			Port:       kubeletPort,
			Protocol:   tcpProtocol,
			SourceCIDR: hostClusterCIDR,
		},
		// Allow traffic from host cluster CIDR to 10300 for node-exporter scraping.
		{
			Port:       nodeExporterPort,
			Protocol:   tcpProtocol,
			SourceCIDR: hostClusterCIDR,
		},
		// Allow traffic from host cluster CIDR to 10301 for kube-state-metrics scraping.
		{
			Port:       kubeStateMetricsPort,
			Protocol:   tcpProtocol,
			SourceCIDR: hostClusterCIDR,
		},
		// Only allow ssh traffic from the host cluster.
		{
			Port:       sshPort,
			Protocol:   tcpProtocol,
			SourceCIDR: hostClusterCIDR,
		},
	}
}

func (s *securityGroupsAdapter) getIngressRules(customObject v1alpha1.AWSConfig) []securityGroupRule {
	return []securityGroupRule{
		// Allow all http and https traffic to the ingress load balancer.
		{
			Port:       httpPort,
			Protocol:   tcpProtocol,
			SourceCIDR: defaultCIDR,
		},
		{
			Port:       httpsPort,
			Protocol:   tcpProtocol,
			SourceCIDR: defaultCIDR,
		},
	}
}

func getKubernetesAPIRule(cfg Config, hostClusterCIDR string) []securityGroupRule {
	// when api whitelisting is enabled, add separate security group rule per each subnet
	if cfg.APIWhitelist.Enabled {
		rules := []securityGroupRule{
			// allow traffic from host cluster CIDR
			{
				Port:       key.KubernetesAPISecurePort(cfg.CustomObject),
				Protocol:   tcpProtocol,
				SourceCIDR: hostClusterCIDR,
			},
			// allow traffic from guest cluster CIDR
			{
				Port:       key.KubernetesAPISecurePort(cfg.CustomObject),
				Protocol:   tcpProtocol,
				SourceCIDR: cfg.CustomObject.Spec.AWS.VPC.CIDR,
			},
		}
		whitelistSubnets := strings.Split(cfg.APIWhitelist.SubnetList, ",")
		for _, subnet := range whitelistSubnets {
			subnetRule := securityGroupRule{
				Port:       key.KubernetesAPISecurePort(cfg.CustomObject),
				Protocol:   tcpProtocol,
				SourceCIDR: subnet,
			}
			rules = append(rules, subnetRule)
		}
		return rules
	} else {
		// when api whitelisting is disabled, allow all traffic
		return []securityGroupRule{
			{
				Port:       key.KubernetesAPISecurePort(cfg.CustomObject),
				Protocol:   tcpProtocol,
				SourceCIDR: defaultCIDR,
			},
		}
	}
}
