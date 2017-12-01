package legacyv1

import (
	"fmt"

	"github.com/giantswarm/microerror"

	awsutil "github.com/giantswarm/aws-operator/client/aws"
	"github.com/giantswarm/aws-operator/resources"
	awsresources "github.com/giantswarm/aws-operator/resources/aws"
)

type securityGroupRulesInput struct {
	Clients   awsutil.Clients
	GroupName string
}

func (s *Resource) deleteSecurityGroupRules(input securityGroupRulesInput) error {
	var securityGroupRules resources.DeletableResource
	securityGroupRules = awsresources.SecurityGroupRules{
		Description: input.GroupName,
		GroupName:   input.GroupName,
		AWSEntity:   awsresources.AWSEntity{Clients: input.Clients},
	}
	if err := securityGroupRules.Delete(); err != nil {
		return microerror.Mask(err)
	}

	s.logger.Log("info", fmt.Sprintf("deleted rules for security group '%s'", input.GroupName))

	return nil
}
