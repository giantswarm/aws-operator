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
	Rules       []SecurityGroupRule
	id          string
	AWSEntity
}

// SecurityGroupRule is an AWS security group rule.
type SecurityGroupRule struct {
	Port int
	// SourceCIDR is the CIDR of the source.
	SourceCIDR string
	// SecurityGroupID is the ID of the source Security Group.
	SecurityGroupID string
}

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

	if len(securityGroups.SecurityGroups) != 1 {
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

// createRule creates a security group rule.
// SourceCIDR always takes precedence over SecurityGroupID.
func (s *SecurityGroup) createRule(rule SecurityGroupRule) error {
	groupID, err := s.GetID()
	if err != nil {
		return microerror.MaskAny(err)
	}

	var params *ec2.AuthorizeSecurityGroupIngressInput
	if rule.SourceCIDR != "" {
		params = &ec2.AuthorizeSecurityGroupIngressInput{
			CidrIp:     aws.String(rule.SourceCIDR),
			GroupId:    aws.String(groupID),
			IpProtocol: aws.String("tcp"),
			FromPort:   aws.Int64(int64(rule.Port)),
			ToPort:     aws.Int64(int64(rule.Port)),
		}
	} else {
		params = &ec2.AuthorizeSecurityGroupIngressInput{
			GroupId: aws.String(groupID),
			IpPermissions: []*ec2.IpPermission{
				{
					FromPort:   aws.Int64(int64(rule.Port)),
					ToPort:     aws.Int64(int64(rule.Port)),
					IpProtocol: aws.String("tcp"),
					UserIdGroupPairs: []*ec2.UserIdGroupPair{
						{
							GroupId: aws.String(rule.SecurityGroupID),
						},
					},
				},
			},
		}
	}

	if _, err := s.Clients.EC2.AuthorizeSecurityGroupIngress(params); err != nil {
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

	s.ApplyRules(s.Rules)

	return nil
}

func (s SecurityGroup) ApplyRules(rules []SecurityGroupRule) error {
	for _, rule := range rules {
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
