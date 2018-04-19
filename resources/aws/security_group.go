package aws

import (
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/cenkalti/backoff"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"

	awsclient "github.com/giantswarm/aws-operator/client/aws"
)

// SecurityGroup is an AWS security group.
type SecurityGroup struct {
	Description string
	GroupName   string
	VpcID       string
	Rules       []SecurityGroupRule
	id          string
	// Dependencies.
	Logger micrologger.Logger
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
	// SecurityGroupID is the ID of the source security Group.
	SecurityGroupID string
}

const (
	allPorts = -1
)

func (s SecurityGroup) findExisting() (*ec2.SecurityGroup, error) {
	filters, err := s.getGroupFilters()
	if err != nil {
		return nil, microerror.Mask(err)
	}

	securityGroups, err := s.Clients.EC2.DescribeSecurityGroups(&ec2.DescribeSecurityGroupsInput{
		Filters: filters,
	})
	if err != nil {
		return nil, microerror.Mask(err)
	}

	if len(securityGroups.SecurityGroups) < 1 {
		return nil, microerror.Maskf(notFoundError, notFoundErrorFormat, SecurityGroupType, s.GroupName)
	} else if len(securityGroups.SecurityGroups) > 1 {
		return nil, microerror.Mask(tooManyResultsError)
	}

	return securityGroups.SecurityGroups[0], nil
}

// findGroupWithRule checks if a security group exists with the specified name,
// description and rule. SourceCIDR always takes precedence over SecurityGroupID.
func (s SecurityGroup) findGroupWithRule(rule SecurityGroupRule) (*ec2.SecurityGroup, error) {
	filters, err := s.getGroupFilters()
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var ruleFilters []*ec2.Filter

	if rule.SourceCIDR != "" {
		ruleFilters = []*ec2.Filter{
			{
				Name: aws.String(ipPermissionCIDR),
				Values: []*string{
					aws.String(rule.SourceCIDR),
				},
			},
			{
				Name: aws.String(ipPermissionFromPort),
				Values: []*string{
					aws.String(strconv.Itoa(rule.Port)),
				},
			},
			{
				Name: aws.String(ipPermissionProtocol),
				Values: []*string{
					aws.String(rule.Protocol),
				},
			},
			{
				Name: aws.String(ipPermissionToPort),
				Values: []*string{
					aws.String(strconv.Itoa(rule.Port)),
				},
			},
		}
	} else {
		ruleFilters = []*ec2.Filter{
			{
				Name: aws.String(ipPermissionGroupID),
				Values: []*string{
					aws.String(rule.SecurityGroupID),
				},
			},
			{
				Name: aws.String(ipPermissionProtocol),
				Values: []*string{
					aws.String(rule.Protocol),
				},
			},
		}

		// Needed as filters for all ports rules don't work.
		if rule.Port != allPorts {
			portFilter := &ec2.Filter{
				Name: aws.String(ipPermissionFromPort),
				Values: []*string{
					aws.String(strconv.Itoa(rule.Port)),
				},
			}

			ruleFilters = append(ruleFilters, portFilter)
		}
	}

	for _, filter := range ruleFilters {
		filters = append(filters, filter)
	}

	securityGroups, err := s.Clients.EC2.DescribeSecurityGroups(&ec2.DescribeSecurityGroupsInput{
		Filters: filters,
	})
	if err != nil {
		return nil, microerror.Mask(err)
	}

	if len(securityGroups.SecurityGroups) < 1 {
		return nil, nil
	} else if len(securityGroups.SecurityGroups) > 1 {
		return nil, microerror.Mask(tooManyResultsError)
	}

	return securityGroups.SecurityGroups[0], nil
}

// CreateIfNotExists creates the security group if it does not exist.
func (s *SecurityGroup) CreateIfNotExists() (bool, error) {
	if err := s.CreateOrFail(); err != nil {
		if strings.Contains(err.Error(), awsclient.SecurityGroupDuplicate) {
			securityGroup, err := s.findExisting()
			if err != nil {
				return false, microerror.Mask(err)
			}
			s.id = *securityGroup.GroupId

			return false, nil
		}

		return false, microerror.Mask(err)
	}

	return true, nil
}

// createRule creates a security group rule.
// SourceCIDR always takes precedence over SecurityGroupID.
func (s *SecurityGroup) createRuleIfNotExists(rule SecurityGroupRule) (bool, error) {
	groupID, err := s.GetID()
	if err != nil {
		return false, microerror.Mask(err)
	}

	existingGroup, err := s.findGroupWithRule(rule)
	if err != nil {
		return false, microerror.Mask(err)
	}
	if existingGroup != nil {
		return true, nil
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
		return false, microerror.Mask(err)
	}

	return true, nil
}

// CreateOrFail creates the security group or returns an error.
func (s *SecurityGroup) CreateOrFail() error {
	securityGroup, err := s.Clients.EC2.CreateSecurityGroup(&ec2.CreateSecurityGroupInput{
		Description: aws.String(s.Description),
		GroupName:   aws.String(s.GroupName),
		VpcId:       aws.String(s.VpcID),
	})
	if err != nil {
		return microerror.Mask(err)
	}

	s.id = *securityGroup.GroupId

	s.ApplyRules(s.Rules)

	return nil
}

// ApplyRules creates the security group rules.
func (s SecurityGroup) ApplyRules(rules []SecurityGroupRule) error {
	for _, rule := range rules {
		if _, err := s.createRuleIfNotExists(rule); err != nil {
			return microerror.Mask(err)
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
		return microerror.Mask(err)
	}

	deleteOperation := func() error {
		if _, err := s.Clients.EC2.DeleteSecurityGroup(&ec2.DeleteSecurityGroupInput{
			GroupId: securityGroup.GroupId,
		}); err != nil {
			return microerror.Mask(err)
		}
		return nil
	}

	deleteNotify := NewNotify(s.Logger, "deleting security group")
	if err := backoff.RetryNotify(deleteOperation, NewCustomExponentialBackoff(), deleteNotify); err != nil {
		return microerror.Mask(err)
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
		return "", microerror.Mask(err)
	}

	return *securityGroup.GroupId, nil
}

func (s SecurityGroup) getGroupFilters() ([]*ec2.Filter, error) {
	if s.Description == "" {
		return nil, microerror.Maskf(attributeEmptyError, attributeEmptyErrorFormat, "Description")
	}
	if s.GroupName == "" {
		return nil, microerror.Maskf(attributeEmptyError, attributeEmptyErrorFormat, "GroupName")
	}

	filters := []*ec2.Filter{
		{
			Name: aws.String(subnetDescription),
			Values: []*string{
				aws.String(s.Description),
			},
		},
		{
			Name: aws.String(subnetGroupName),
			Values: []*string{
				aws.String(s.GroupName),
			},
		},
	}

	return filters, nil
}
