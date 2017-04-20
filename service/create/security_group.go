package create

import (
	"fmt"

	"github.com/giantswarm/awstpr"
	"github.com/juju/errgo"

	awsutil "github.com/giantswarm/aws-operator/client/aws"
	"github.com/giantswarm/aws-operator/resources"
	awsresources "github.com/giantswarm/aws-operator/resources/aws"
)

type securityGroupInput struct {
	Clients           awsutil.Clients
	Cluster           awstpr.CustomObject
	PortsToOpen       []int
	SecurityGroupType string
	VpcID             string
}

func (s *Service) createSecurityGroup(input securityGroupInput) (resources.ResourceWithID, error) {
	groupName := securityGroupName(input)

	securityGroup := &awsresources.SecurityGroup{
		Description: groupName,
		GroupName:   groupName,
		VpcID:       input.VpcID,
		PortsToOpen: input.PortsToOpen,
		AWSEntity:   awsresources.AWSEntity{Clients: input.Clients},
	}
	securityGroupCreated, err := securityGroup.CreateIfNotExists()
	if err != nil {
		s.logger.Log("error", fmt.Sprintf("could not create security group '%s': %s", groupName, errgo.Details(err)))
		return nil, err
	}
	if securityGroupCreated {
		s.logger.Log("info", fmt.Sprintf("created security group '%s'", groupName))
	} else {
		s.logger.Log("info", fmt.Sprintf("security group '%s' already exists, reusing", groupName))
	}

	return securityGroup, nil
}

func (s *Service) deleteSecurityGroup(input securityGroupInput) error {
	groupName := securityGroupName(input)

	var securityGroup resources.ResourceWithID
	securityGroup = &awsresources.SecurityGroup{
		Description: groupName,
		GroupName:   groupName,
		AWSEntity:   awsresources.AWSEntity{Clients: input.Clients},
	}
	if err := securityGroup.Delete(); err != nil {
		s.logger.Log("error", fmt.Sprintf("could not delete security group '%s': %s", groupName, errgo.Details(err)))
		return err
	} else {
		s.logger.Log("info", "deleted security group '%s'", groupName)
		return nil
	}
}

func securityGroupName(input securityGroupInput) string {
	return fmt.Sprintf("%s-%s", input.Cluster.Name, input.SecurityGroupType)
}
