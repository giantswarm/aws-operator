package aws

import (
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	microerror "github.com/giantswarm/microkit/error"
)

type SecurityGroup struct {
	Description string
	GroupName   string
	VpcID       string
	PortsToOpen []int
	id          string
	AWSEntity
}

func (s SecurityGroup) findExisting() (*ec2.SecurityGroup, error) {
	portRuleFilters := []*ec2.Filter{
		&ec2.Filter{
			Name: aws.String(subnetVpcID),
			Values: []*string{
				aws.String(s.VpcID),
			},
		},
	}

	for _, port := range s.PortsToOpen {
		portRuleFilters = append(portRuleFilters, &ec2.Filter{
			Name: aws.String(securityGroupIPPermissionFromPort),
			Values: []*string{
				aws.String(strconv.Itoa(port)),
			},
		}, &ec2.Filter{
			Name: aws.String(securityGroupIPPermissionToPort),
			Values: []*string{
				aws.String(strconv.Itoa(port)),
			},
		})
	}

	filters := [][]*ec2.Filter{
		portRuleFilters,
		{
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
	}

	for _, filter := range filters {
		securityGroups, err := s.Clients.EC2.DescribeSecurityGroups(&ec2.DescribeSecurityGroupsInput{
			Filters: filter,
		})
		if err != nil {
			return nil, microerror.MaskAny(err)
		}

		if len(securityGroups.SecurityGroups) < 1 {
			continue
		}

		return securityGroups.SecurityGroups[0], nil
	}

	return nil, microerror.MaskAny(securityGroupFindError)
}

func (s *SecurityGroup) checkIfExists() (bool, error) {
	_, err := s.findExisting()
	if IsSecurityGroupFind(err) {
		return false, nil
	} else if err != nil {
		return false, microerror.MaskAny(err)
	}

	return true, nil
}

func (s *SecurityGroup) CreateIfNotExists() (bool, error) {
	exists, err := s.checkIfExists()
	if err != nil {
		return false, microerror.MaskAny(err)
	}

	if exists {
		return false, nil
	}

	if err := s.CreateOrFail(); err != nil {
		return false, microerror.MaskAny(err)
	}

	return true, nil
}

func (s *SecurityGroup) openPort(port int) error {
	groupID, err := s.GetID()
	if err != nil {
		return microerror.MaskAny(err)
	}

	if _, err := s.Clients.EC2.AuthorizeSecurityGroupIngress(&ec2.AuthorizeSecurityGroupIngressInput{
		CidrIp:     aws.String("0.0.0.0/0"),
		GroupId:    aws.String(groupID),
		IpProtocol: aws.String("tcp"),
		FromPort:   aws.Int64(int64(port)),
		ToPort:     aws.Int64(int64(port)),
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

	for _, port := range s.PortsToOpen {
		if err := s.openPort(port); err != nil {
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
