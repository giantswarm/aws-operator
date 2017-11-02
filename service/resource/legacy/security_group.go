package legacy

import (
	"fmt"

	"github.com/giantswarm/awstpr"
	"github.com/giantswarm/microerror"

	awsutil "github.com/giantswarm/aws-operator/client/aws"
	"github.com/giantswarm/aws-operator/resources"
	awsresources "github.com/giantswarm/aws-operator/resources/aws"
	"github.com/giantswarm/aws-operator/service/key"
)

type securityGroupInput struct {
	Clients   awsutil.Clients
	GroupName string
	VPCID     string
}

type rulesInput struct {
	Cluster                awstpr.CustomObject
	Rules                  []awsresources.SecurityGroupRule
	MastersSecurityGroupID string
	WorkersSecurityGroupID string
	IngressSecurityGroupID string
	HostClusterCIDR        string
}

const (
	allPorts  = -1
	httpPort  = 80
	httpsPort = 443
	sshPort   = 22

	allProtocols = "-1"
	tcpProtocol  = "tcp"

	defaultCIDR = "0.0.0.0/0"
)

func (s *Service) createSecurityGroup(input securityGroupInput) (*awsresources.SecurityGroup, error) {
	securityGroup := &awsresources.SecurityGroup{
		Description: input.GroupName,
		GroupName:   input.GroupName,
		VpcID:       input.VPCID,
		AWSEntity:   awsresources.AWSEntity{Clients: input.Clients},
	}
	securityGroupCreated, err := securityGroup.CreateIfNotExists()
	if err != nil {
		return nil, microerror.Mask(err)
	}
	if securityGroupCreated {
		s.logger.Log("info", fmt.Sprintf("created security group '%s'", input.GroupName))
	} else {
		s.logger.Log("info", fmt.Sprintf("security group '%s' already exists, reusing", input.GroupName))
	}

	return securityGroup, nil
}

func (s *Service) deleteSecurityGroup(input securityGroupInput) error {
	var securityGroup resources.ResourceWithID
	securityGroup = &awsresources.SecurityGroup{
		Description: input.GroupName,
		GroupName:   input.GroupName,
		AWSEntity:   awsresources.AWSEntity{Clients: input.Clients},
	}
	if err := securityGroup.Delete(); err != nil {
		return microerror.Mask(err)
	}
	s.logger.Log("info", fmt.Sprintf("deleted security group '%s'", input.GroupName))

	return nil
}

// masterRules returns the rules for the masters security group.
func (ri rulesInput) masterRules() []awsresources.SecurityGroupRule {
	rules := []awsresources.SecurityGroupRule{
		{
			Port:       ri.Cluster.Spec.Cluster.Kubernetes.API.SecurePort,
			Protocol:   tcpProtocol,
			SourceCIDR: defaultCIDR,
		},
		// Allow all traffic between the masters and worker nodes for Calico.
		{
			Port:            allPorts,
			Protocol:        allProtocols,
			SecurityGroupID: ri.MastersSecurityGroupID,
		},
		{
			Port:            allPorts,
			Protocol:        allProtocols,
			SecurityGroupID: ri.WorkersSecurityGroupID,
		},
	}

	if key.HasClusterVersion(ri.Cluster) {
		// For new clusters we only allow SSH access from the host cluster.
		rules = append(rules, awsresources.SecurityGroupRule{
			Port:       sshPort,
			Protocol:   tcpProtocol,
			SourceCIDR: ri.HostClusterCIDR,
		})
	} else {
		// We need to use the default CIDR for SSH on old clusters.
		rules = append(rules, awsresources.SecurityGroupRule{
			Port:       sshPort,
			Protocol:   tcpProtocol,
			SourceCIDR: defaultCIDR,
		})

		// We need to add an etcd rule for old clusters.
		rules = append(rules, awsresources.SecurityGroupRule{
			Port:       ri.Cluster.Spec.Cluster.Etcd.Port,
			Protocol:   tcpProtocol,
			SourceCIDR: defaultCIDR,
		})
	}

	return rules
}

// workerRules returns the rules for the workers security group.
func (ri rulesInput) workerRules() []awsresources.SecurityGroupRule {
	rules := []awsresources.SecurityGroupRule{
		{
			Port:            ri.Cluster.Spec.Cluster.Kubernetes.IngressController.SecurePort,
			Protocol:        tcpProtocol,
			SecurityGroupID: ri.IngressSecurityGroupID,
		},
		{
			Port:            ri.Cluster.Spec.Cluster.Kubernetes.IngressController.InsecurePort,
			Protocol:        tcpProtocol,
			SecurityGroupID: ri.IngressSecurityGroupID,
		},
		// Allow all traffic between the masters and worker nodes for Calico.
		{
			Port:            allPorts,
			Protocol:        allProtocols,
			SecurityGroupID: ri.MastersSecurityGroupID,
		},
		{
			Port:            allPorts,
			Protocol:        allProtocols,
			SecurityGroupID: ri.WorkersSecurityGroupID,
		},
		// Allow traffic from host cluster to ingress controller secure port,
		// for guest cluster scraping.
		{
			Port:       ri.Cluster.Spec.Cluster.Kubernetes.IngressController.SecurePort,
			Protocol:   tcpProtocol,
			SourceCIDR: ri.HostClusterCIDR,
		},
	}

	if key.HasClusterVersion(ri.Cluster) {
		// For new clusters we only allow SSH access from the host cluster.
		rules = append(rules, awsresources.SecurityGroupRule{
			Port:       sshPort,
			Protocol:   tcpProtocol,
			SourceCIDR: ri.HostClusterCIDR,
		})
	} else {
		// We need to use the default CIDR for SSH on old clusters.
		rules = append(rules, awsresources.SecurityGroupRule{
			Port:       sshPort,
			Protocol:   tcpProtocol,
			SourceCIDR: defaultCIDR,
		})
	}

	return rules
}

// ingressRules returns the rules for the ingress ELB security group.
func (ri rulesInput) ingressRules() []awsresources.SecurityGroupRule {
	return []awsresources.SecurityGroupRule{
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
