package create

import (
	"fmt"

	"github.com/giantswarm/awstpr"
	microerror "github.com/giantswarm/microkit/error"

	awsutil "github.com/giantswarm/aws-operator/client/aws"
	"github.com/giantswarm/aws-operator/resources"
	awsresources "github.com/giantswarm/aws-operator/resources/aws"
)

type securityGroupInput struct {
	Clients   awsutil.Clients
	GroupName string
	VPCID     string
}

type rulesInput struct {
	Cluster                awstpr.CustomObject
	Rules                  []awsresources.Rule
	OwnSecurityGroupID     string
	MastersSecurityGroupID string
	IngressSecurityGroupID string
}

const (
	calicoBGPNetworkPort = 179
	httpPort             = 80
	httpsPort            = 443
	sshPort              = 22

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
		return nil, microerror.MaskAny(err)
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
		return microerror.MaskAny(err)
	} else {
		s.logger.Log("info", fmt.Sprintf("deleted security group '%s'", input.GroupName))
	}

	return nil
}

// masterRules returns the rules for the masters security group.
func (ri rulesInput) masterRules() []awsresources.Rule {
	return []awsresources.Rule{
		{
			Port:       ri.Cluster.Spec.Cluster.Kubernetes.API.SecurePort,
			SourceCIDR: defaultCIDR,
		},
		{
			Port:            ri.Cluster.Spec.Cluster.Etcd.Port,
			SecurityGroupID: ri.OwnSecurityGroupID,
		},
		{
			Port:       sshPort,
			SourceCIDR: defaultCIDR,
		},
		{
			Port:       calicoBGPNetworkPort,
			SourceCIDR: defaultCIDR,
		},
	}
}

// workerRules returns the rules for the workers security group.
func (ri rulesInput) workerRules() []awsresources.Rule {
	return []awsresources.Rule{
		{
			Port:            ri.Cluster.Spec.Cluster.Kubernetes.IngressController.InsecurePort,
			SecurityGroupID: ri.IngressSecurityGroupID,
		},
		{
			Port:            ri.Cluster.Spec.Cluster.Kubernetes.IngressController.SecurePort,
			SecurityGroupID: ri.IngressSecurityGroupID,
		},
		{
			Port:            ri.Cluster.Spec.Cluster.Kubernetes.Kubelet.Port,
			SecurityGroupID: ri.MastersSecurityGroupID,
		},
		{
			Port:       sshPort,
			SourceCIDR: defaultCIDR,
		},
		{
			Port:       calicoBGPNetworkPort,
			SourceCIDR: defaultCIDR,
		},
	}
}

// ingressRules returns the rules for the ingress security group.
func (ri rulesInput) ingressRules() []awsresources.Rule {
	return []awsresources.Rule{
		{
			Port:       httpPort,
			SourceCIDR: defaultCIDR,
		},
		{
			Port:       httpsPort,
			SourceCIDR: defaultCIDR,
		},
		{
			Port:       calicoBGPNetworkPort,
			SourceCIDR: defaultCIDR,
		},
	}
}

func securityGroupName(clusterName string, groupName string) string {
	return fmt.Sprintf("%s-%s", clusterName, groupName)
}
