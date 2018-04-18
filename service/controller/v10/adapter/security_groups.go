package adapter

import (
	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/service/controller/v10/key"
)

// The template related to this adapter can be found in the following import.
//
//     github.com/giantswarm/aws-operator/service/controller/v10/templates/cloudformation/guest/security_groups.go
//

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
	s.MasterSecurityGroupRules = s.getMasterRules(cfg.CustomObject, hostClusterCIDR)

	s.WorkerSecurityGroupName = key.SecurityGroupName(cfg.CustomObject, prefixWorker)
	s.WorkerSecurityGroupRules = s.getWorkerRules(cfg.CustomObject, hostClusterCIDR)

	s.IngressSecurityGroupName = key.SecurityGroupName(cfg.CustomObject, prefixIngress)
	s.IngressSecurityGroupRules = s.getIngressRules(cfg.CustomObject)

	return nil
}

func (s *securityGroupsAdapter) getMasterRules(customObject v1alpha1.AWSConfig, hostClusterCIDR string) []securityGroupRule {
	return []securityGroupRule{
		// Allow all traffic to the kubernetes api server.
		{
			Port:       key.KubernetesAPISecurePort(customObject),
			Protocol:   tcpProtocol,
			SourceCIDR: defaultCIDR,
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
