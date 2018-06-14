package adapter

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/service/controller/v13/key"
)

// The template related to this adapter can be found in the following import.
//
//     github.com/giantswarm/aws-operator/service/controller/v13/templates/cloudformation/guest/security_groups.go
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

func (s *securityGroupsAdapter) getSecurityGroups(cfg Config) error {
	hostClusterCIDR, err := VpcCIDR(cfg.HostClients, key.PeerID(cfg.CustomObject))
	if err != nil {
		return microerror.Mask(err)
	}

	masterRules, err := s.getMasterRules(cfg, hostClusterCIDR)
	if err != nil {
		return microerror.Mask(err)
	}

	s.MasterSecurityGroupName = key.SecurityGroupName(cfg.CustomObject, prefixMaster)
	s.MasterSecurityGroupRules = masterRules

	s.WorkerSecurityGroupName = key.SecurityGroupName(cfg.CustomObject, prefixWorker)
	s.WorkerSecurityGroupRules = s.getWorkerRules(cfg.CustomObject, hostClusterCIDR)

	s.IngressSecurityGroupName = key.SecurityGroupName(cfg.CustomObject, prefixIngress)
	s.IngressSecurityGroupRules = s.getIngressRules(cfg.CustomObject)

	return nil
}

func (s *securityGroupsAdapter) getMasterRules(cfg Config, hostClusterCIDR string) ([]securityGroupRule, error) {
	// Allow traffic to the Kubernetes API server depending on the API
	// whitelisting rules.
	apiRules, err := getKubernetesAPIRules(cfg, hostClusterCIDR)
	if err != nil {
		return []securityGroupRule{}, microerror.Mask(err)
	}

	// Other security group rules for the master.
	otherRules := []securityGroupRule{
		// Allow traffic from host cluster CIDR to 4194 for cadvisor scraping.
		{
			Port:       cadvisorPort,
			Protocol:   tcpProtocol,
			SourceCIDR: hostClusterCIDR,
		},
		// Allow traffic from host cluster CIDR to 2379 for etcd backup.
		{
			Port:       etcdPort,
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
	return append(apiRules, otherRules...), nil
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

func getKubernetesAPIRules(cfg Config, hostClusterCIDR string) ([]securityGroupRule, error) {
	// When API whitelisting is enabled, add separate security group rule per each subnet.
	if cfg.APIWhitelist.Enabled {
		rules := []securityGroupRule{
			// Allow traffic from host cluster CIDR.
			{
				Port:       key.KubernetesAPISecurePort(cfg.CustomObject),
				Protocol:   tcpProtocol,
				SourceCIDR: hostClusterCIDR,
			},
			// Allow traffic from guest cluster CIDR.
			{
				Port:       key.KubernetesAPISecurePort(cfg.CustomObject),
				Protocol:   tcpProtocol,
				SourceCIDR: cfg.CustomObject.Spec.AWS.VPC.CIDR,
			},
		}

		// Whitelist all configured subnets.
		whitelistSubnets := strings.Split(cfg.APIWhitelist.SubnetList, ",")
		for _, subnet := range whitelistSubnets {
			if subnet != "" {
				subnetRule := securityGroupRule{
					Port:       key.KubernetesAPISecurePort(cfg.CustomObject),
					Protocol:   tcpProtocol,
					SourceCIDR: subnet,
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
				Port:       key.KubernetesAPISecurePort(cfg.CustomObject),
				Protocol:   tcpProtocol,
				SourceCIDR: defaultCIDR,
			},
		}

		return allowAllRule, nil
	}
}

func getHostClusterNATGatewayRules(cfg Config) ([]securityGroupRule, error) {
	gatewayRules := []securityGroupRule{}

	// Get all EIPs tagged with the host cluster installation tag.
	// Each EIP is associated with a host cluster NAT gateway.
	describeAddressesInput := &ec2.DescribeAddressesInput{
		Filters: []*ec2.Filter{
			{
				Name: aws.String("tag:giantswarm.io/installation"),
				Values: []*string{
					aws.String(cfg.InstallationName),
				},
			},
		},
	}
	output, err := cfg.HostClients.EC2.DescribeAddresses(describeAddressesInput)
	if err != nil {
		return gatewayRules, microerror.Mask(err)
	}

	for _, address := range output.Addresses {
		gatewayRule := securityGroupRule{
			Port:       key.KubernetesAPISecurePort(cfg.CustomObject),
			Protocol:   tcpProtocol,
			SourceCIDR: fmt.Sprintf("%s/32", *address.PublicIp),
		}

		gatewayRules = append(gatewayRules, gatewayRule)
	}

	return gatewayRules, nil
}
