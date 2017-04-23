package create

import (
	"fmt"

	microerror "github.com/giantswarm/microkit/error"

	awsutil "github.com/giantswarm/aws-operator/client/aws"
	"github.com/giantswarm/aws-operator/resources"
	awsresources "github.com/giantswarm/aws-operator/resources/aws"
)

type securityGroupInput struct {
	Clients     awsutil.Clients
	GroupName   string
	PortsToOpen []int
	VPCID       string
}

func (s *Service) createSecurityGroup(input securityGroupInput) (resources.ResourceWithID, error) {

	var securityGroup resources.ResourceWithID
	securityGroup = &awsresources.SecurityGroup{
		Description: input.GroupName,
		GroupName:   input.GroupName,
		VpcID:       input.VPCID,
		PortsToOpen: input.PortsToOpen,
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
		s.logger.Log("info", "deleted security group '%s'", input.GroupName)
	}

	return nil
}

func securityGroupName(clusterName string, groupName string) string {
	return fmt.Sprintf("%s-%s", clusterName, groupName)
}
