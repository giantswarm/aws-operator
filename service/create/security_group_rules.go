package create

import (
	"fmt"

	microerror "github.com/giantswarm/microkit/error"

	awsutil "github.com/giantswarm/aws-operator/client/aws"
	"github.com/giantswarm/aws-operator/resources"
	awsresources "github.com/giantswarm/aws-operator/resources/aws"
)

type securityGroupRulesInput struct {
	Clients   awsutil.Clients
	GroupName string
}

func (s *Service) deleteSecurityGroupRules(input securityGroupRulesInput) error {
	var securityGroupRules resources.DeletableResource
	securityGroupRules = awsresources.SecurityGroupRules{
		Description: input.GroupName,
		GroupName:   input.GroupName,
		AWSEntity:   awsresources.AWSEntity{Clients: input.Clients},
	}
	if err := securityGroupRules.Delete(); err != nil {
		return microerror.MaskAny(err)
	}

	s.logger.Log("info", fmt.Sprintf("deleted rules for security group '%s'", input.GroupName))

	return nil
}
