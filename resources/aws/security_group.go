package aws

import (
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	awsclient "github.com/giantswarm/aws-operator/client/aws"
	microerror "github.com/giantswarm/microkit/error"
)

type SecurityGroup struct {
	Description string
	GroupName   string
	VpcID       string
	Rules       Rules
	id          string
	AWSEntity
}

// Rule is a Security Group rule.
type Rule struct {
	Port int
	// SourceCIDR is the CIDR of the source. Conflicts with SourceCIDR.
	SourceCIDR string
	// SecurityGroupID is the ID of the Security Group. Conflicts with SecurityGroupID.
	SecurityGroupID string
}

// Rules is a slice of Rule structs.
type Rules []Rule

func (s SecurityGroup) findExisting() (*ec2.SecurityGroup, error) {
	securityGroups, err := s.Clients.EC2.DescribeSecurityGroups(&ec2.DescribeSecurityGroupsInput{
		Filters: []*ec2.Filter{
			&ec2.Filter{
				Name: aws.String(subnetDescription),
				Values: []*string{
					aws.String(s.Description),
				},
			},
			&ec2.Filter{
				Name: aws.String(subnetGroupName),
				Values: []*string{
					aws.String(s.GroupName),
				},
			},
		},
	})
	if err != nil {
		return nil, microerror.MaskAny(err)
	}

	if len(securityGroups.SecurityGroups) < 1 {
		return nil, microerror.MaskAny(securityGroupFindError)
	}

	return securityGroups.SecurityGroups[0], nil
}

func (s *SecurityGroup) CreateIfNotExists() (bool, error) {
	if err := s.CreateOrFail(); err != nil {
		if strings.Contains(err.Error(), awsclient.SecurityGroupDuplicate) {
			securityGroup, err := s.findExisting()
			if err != nil {
				return false, microerror.MaskAny(err)
			}
			s.id = *securityGroup.GroupId

			return false, nil
		}

		return false, microerror.MaskAny(err)
	}

	return true, nil
}

func (s *SecurityGroup) createRule(rule Rule) error {
	groupID, err := s.GetID()
	if err != nil {
		return microerror.MaskAny(err)
	}

	if _, err := s.Clients.EC2.AuthorizeSecurityGroupIngress(&ec2.AuthorizeSecurityGroupIngressInput{
		CidrIp:     aws.String(rule.SourceCIDR),
		GroupId:    aws.String(groupID),
		IpProtocol: aws.String("tcp"),
		FromPort:   aws.Int64(int64(rule.Port)),
		ToPort:     aws.Int64(int64(rule.Port)),
	}); err != nil {
		return microerror.MaskAny(err)
	}

	return nil
}

func (s *SecurityGroup) CreateOrFail() error {
	securityGroup, err := s.Clients.EC2.CreateSecurityGroup(&ec2.CreateSecurityGroupInput{
		Description: aws.String(s.Description),
		GroupName:   aws.String(s.GroupName),
		VpcId:       aws.String(s.VpcID),
	})
	if err != nil {
		return microerror.MaskAny(err)
	}

	s.id = *securityGroup.GroupId

	for _, rule := range s.Rules {
		if err := s.createRule(rule); err != nil {
			return microerror.MaskAny(err)
		}
	}

	return nil
}

func (s *SecurityGroup) Delete() error {
	securityGroup, err := s.findExisting()
	if err != nil {
		return microerror.MaskAny(err)
	}

	if _, err := s.Clients.EC2.DeleteSecurityGroup(&ec2.DeleteSecurityGroupInput{
		GroupId: securityGroup.GroupId,
	}); err != nil {
		return microerror.MaskAny(err)
	}

	return nil
}

func (s SecurityGroup) GetID() (string, error) {
	if s.id != "" {
		return s.id, nil
	}

	securityGroup, err := s.findExisting()
	if err != nil {
		return "", microerror.MaskAny(err)
	}

	return *securityGroup.GroupId, nil
}
