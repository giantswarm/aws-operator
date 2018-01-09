package adapter

import (
	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/service/keyv2"
)

// template related to this adapter: service/templates/cloudformation/security_groups.yaml

type securityGroupsAdapter struct {
	MasterSecurityGroupName   string
	WorkerSecurityGroupName   string
	WorkerSecurityGroupRules  []securityGroupRule
	IngressSecurityGroupName  string
	IngressSecurityGroupRules []securityGroupRule
}

type securityGroupRule struct {
	Port                    int
	Protocol                string
	SourceCIDR              string
	SourceSecurityGroupName string
}

const (
	allPorts         = -1
	kubeletPort      = 10250
	nodeExporterPort = 10300
	sshPort          = 22

	allProtocols = "-1"
	tcpProtocol  = "tcp"

	defaultCIDR = "0.0.0.0/0"
)

func (s *securityGroupsAdapter) getSecurityGroups(customObject v1alpha1.AWSConfig, clients Clients) error {
	hostClusterCIDR, err := VpcCIDR(clients, customObject.Spec.AWS.VPC.PeerID)
	if err != nil {
		return microerror.Mask(err)
	}

	s.MasterSecurityGroupName = keyv2.SecurityGroupName(customObject, prefixMaster)
	s.WorkerSecurityGroupName = keyv2.SecurityGroupName(customObject, prefixWorker)
	s.WorkerSecurityGroupRules = s.getWorkerRules(customObject, hostClusterCIDR)

	s.IngressSecurityGroupName = keyv2.SecurityGroupName(customObject, prefixIngress)
	s.IngressSecurityGroupRules = s.getIngressRules(customObject)

	return nil
}

func (s *securityGroupsAdapter) getWorkerRules(customObject v1alpha1.AWSConfig, hostClusterCIDR string) []securityGroupRule {
	return []securityGroupRule{
		{
			Port:                    customObject.Spec.Cluster.Kubernetes.IngressController.SecurePort,
			Protocol:                tcpProtocol,
			SourceSecurityGroupName: keyv2.SecurityGroupName(customObject, prefixIngress),
		},
		{
			Port:                    customObject.Spec.Cluster.Kubernetes.IngressController.InsecurePort,
			Protocol:                tcpProtocol,
			SourceSecurityGroupName: keyv2.SecurityGroupName(customObject, prefixIngress),
		},
		// Allow all traffic between the masters and worker nodes for Calico.
		{
			Port:                    allPorts,
			Protocol:                allProtocols,
			SourceSecurityGroupName: keyv2.SecurityGroupName(customObject, prefixMaster),
		},
		{
			Port:                    allPorts,
			Protocol:                allProtocols,
			SourceSecurityGroupName: keyv2.SecurityGroupName(customObject, prefixWorker),
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
