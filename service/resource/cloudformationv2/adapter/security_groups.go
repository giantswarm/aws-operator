package adapter

import (
	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/service/keyv2"
)

// template related to this adapter: service/templates/cloudformation/guest/security_groups.yaml

type securityGroupsAdapter struct {
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
	allPorts         = -1
	kubeletPort      = 10250
	nodeExporterPort = 10300
	sshPort          = 22

	allProtocols = "-1"
	tcpProtocol  = "tcp"

	defaultCIDR = "0.0.0.0/0"

	ingressSecurityGroupName = "IngressSecurityGroup"
)

func (s *securityGroupsAdapter) getSecurityGroups(cfg Config) error {
	hostClusterCIDR, err := VpcCIDR(cfg.HostClients, cfg.CustomObject.Spec.AWS.VPC.PeerID)
	if err != nil {
		return microerror.Mask(err)
	}

	s.MasterSecurityGroupName = keyv2.SecurityGroupName(cfg.CustomObject, prefixMaster)
	s.MasterSecurityGroupRules = s.getMasterRules(cfg.CustomObject, hostClusterCIDR)

	s.WorkerSecurityGroupName = keyv2.SecurityGroupName(cfg.CustomObject, prefixWorker)
	s.WorkerSecurityGroupRules = s.getWorkerRules(cfg.CustomObject, hostClusterCIDR)

	s.IngressSecurityGroupName = keyv2.SecurityGroupName(cfg.CustomObject, prefixIngress)
	s.IngressSecurityGroupRules = s.getIngressRules(cfg.CustomObject)

	return nil
}

func (s *securityGroupsAdapter) getMasterRules(customObject v1alpha1.AWSConfig, hostClusterCIDR string) []securityGroupRule {
	return []securityGroupRule{
		// Allow all traffic to the kubernetes api server.
		{
			Port:       keyv2.KubernetesAPISecurePort(customObject),
			Protocol:   tcpProtocol,
			SourceCIDR: defaultCIDR,
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
		// Only allow ssh traffic from the host cluster.
		{
			Port:       sshPort,
			Protocol:   tcpProtocol,
			SourceCIDR: hostClusterCIDR,
		},
	}
}

func (s *securityGroupsAdapter) getWorkerRules(customObject v1alpha1.AWSConfig, hostClusterCIDR string) []securityGroupRule {
	return []securityGroupRule{
		// Allow traffic from the ingress security group to the ingress controller.
		{
			Port:                keyv2.IngressControllerSecurePort(customObject),
			Protocol:            tcpProtocol,
			SourceSecurityGroup: ingressSecurityGroupName,
		},
		{
			Port:                keyv2.IngressControllerInsecurePort(customObject),
			Protocol:            tcpProtocol,
			SourceSecurityGroup: ingressSecurityGroupName,
		},
		// Allow traffic from host cluster to ingress controller secure port,
		// for guest cluster scraping.
		{
			Port:       keyv2.IngressControllerSecurePort(customObject),
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
