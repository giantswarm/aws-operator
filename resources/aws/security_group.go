package aws

import (
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	awsclient "github.com/giantswarm/aws-operator/client/aws"
	microerror "github.com/giantswarm/microkit/error"
)

// SecurityGroup is an AWS security group.
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
	// Port is the port to open.
	Port int
	// Protocol is the IP protocol.
	Protocol string
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

	if len(securityGroups.SecurityGroups) < 1 {
		return nil, microerror.MaskAnyf(notFoundError, notFoundErrorFormat, SecurityGroupType, s.GroupName)
	} else if len(securityGroups.SecurityGroups) > 1 {
		return nil, microerror.MaskAny(tooManyResultsError)
	}

	return securityGroups.SecurityGroups[0], nil
}

// CreateIfNotExists creates the security group if it does not exist.
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
			IpProtocol: aws.String(rule.Protocol),
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
					IpProtocol: aws.String(rule.Protocol),
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

// deleteRule creates a security group rule.
// SourceCIDR always takes precedence over SecurityGroupID.
func (s *SecurityGroup) deleteRule(rule SecurityGroupRule) error {
	groupID, err := s.GetID()
	if err != nil {
		return microerror.MaskAny(err)
	}

	var params *ec2.RevokeSecurityGroupEgressInput
	if rule.SourceCIDR != "" {
		params = &ec2.RevokeSecurityGroupEgressInput{
			CidrIp:     aws.String(rule.SourceCIDR),
			GroupId:    aws.String(groupID),
			IpProtocol: aws.String(rule.Protocol),
			FromPort:   aws.Int64(int64(rule.Port)),
			ToPort:     aws.Int64(int64(rule.Port)),
		}
	} else {
		params = &ec2.RevokeSecurityGroupEgressInput{
			GroupId: aws.String(groupID),
			IpPermissions: []*ec2.IpPermission{
				&ec2.IpPermission{
					FromPort:   aws.Int64(int64(rule.Port)),
					ToPort:     aws.Int64(int64(rule.Port)),
					IpProtocol: aws.String(rule.Protocol),
					UserIdGroupPairs: []*ec2.UserIdGroupPair{
						{
							GroupId: aws.String(rule.SecurityGroupID),
						},
					},
				},
			},
		}
	}

	if _, err := s.Clients.EC2.RevokeSecurityGroupEgress(params); err != nil {
		return microerror.MaskAny(err)
	}

	return nil
}

// CreateOrFail creates the security group or returns an error.
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

// ApplyRules creates the security group rules.
func (s SecurityGroup) ApplyRules(rules []SecurityGroupRule) error {
	for _, rule := range rules {
		if err := s.createRule(rule); err != nil {
			return microerror.MaskAny(err)
		}
	}

	return nil
}

// Delete deletes the security group. A security group cannot be deleted if it
// references another securty group. So first we delete any rules referencing
// other groups.
func (s *SecurityGroup) Delete() error {
	securityGroup, err := s.findExisting()
	if err != nil {
		return microerror.MaskAny(err)
	}

	for _, ipPermission := range securityGroup.IpPermissions {
		// Rule references a security group not a CIDR so it must be deleted.
		if len(ipPermission.UserIdGroupPairs) > 0 {
			for _, userIDGroupPair := range ipPermission.UserIdGroupPairs {
				rule := SecurityGroupRule{
					Port:            int(*ipPermission.FromPort),
					Protocol:        *ipPermission.IpProtocol,
					SecurityGroupID: *userIDGroupPair.GroupId,
				}

				if err := s.deleteRule(rule); err != nil {
					return microerror.MaskAny(err)
				}
			}
		}
	}

	if _, err := s.Clients.EC2.DeleteSecurityGroup(&ec2.DeleteSecurityGroupInput{
		GroupId: securityGroup.GroupId,
	}); err != nil {
		return microerror.MaskAny(err)
	}

	return nil
}

// GetID gets the AWS security group ID.
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
