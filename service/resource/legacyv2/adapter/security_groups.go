package adapter

import (
	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"

	"github.com/giantswarm/aws-operator/service/keyv2"
)

// template related to this adapter: service/templates/cloudformation/security_groups.yaml

type securityGroupsAdapter struct {
	MasterGroupName           string
	WorkerGroupName           string
	IngressGroupName          string
	IngressSecurityGroupRules []securityGroupRule
}

type securityGroupRule struct {
	Port              int
	Protocol          string
	SourceCIDR        string
	SecurityGroupName string
}

const (
	allProtocols = "-1"
	tcpProtocol  = "tcp"

	defaultCIDR = "0.0.0.0/0"
)

func (s *securityGroupsAdapter) getSecurityGroups(customObject v1alpha1.AWSConfig, clients Clients) error {
	s.MasterGroupName = keyv2.SecurityGroupName(customObject, prefixMaster)
	s.WorkerGroupName = keyv2.SecurityGroupName(customObject, prefixWorker)

	s.IngressGroupName = keyv2.SecurityGroupName(customObject, prefixIngress)
	s.IngressSecurityGroupRules = s.getIngressRules(customObject)

	return nil
}

func (s *securityGroupsAdapter) getIngressRules(customObject v1alpha1.AWSConfig) []securityGroupRule {
	return []securityGroupRule{
		securityGroupRule{
			Port:       httpPort,
			Protocol:   tcpProtocol,
			SourceCIDR: defaultCIDR,
		},
		securityGroupRule{
			Port:       httpsPort,
			Protocol:   tcpProtocol,
			SourceCIDR: defaultCIDR,
		},
	}
}
