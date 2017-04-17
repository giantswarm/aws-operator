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
	PortsToOpen []int
	id          string
	AWSEntity
}

func (s SecurityGroup) list() ([]*ec2.SecurityGroup, error) {
	out, err := s.Clients.EC2.DescribeSecurityGroups(&ec2.DescribeSecurityGroupsInput{
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
	return out.SecurityGroups, microerror.MaskAny(err)
}

func (s *SecurityGroup) CreateIfNotExists() (bool, error) {
	if err := s.CreateOrFail(); err != nil {
		if strings.Contains(err.Error(), awsclient.SecurityGroupDuplicate) {
			securityGroups, err := s.list()
			if err == nil && len(securityGroups) > 0 {
				s.id = *securityGroups[0].GroupId
			}
			return false, microerror.MaskAny(err)
		}

		return false, microerror.MaskAny(err)
	}

	return true, nil
}

func (s *SecurityGroup) openPort(port int) error {
	if _, err := s.Clients.EC2.AuthorizeSecurityGroupIngress(&ec2.AuthorizeSecurityGroupIngressInput{
		CidrIp:     aws.String("0.0.0.0/0"),
		GroupId:    aws.String(s.ID()),
		IpProtocol: aws.String("tcp"),
		FromPort:   aws.Int64(0),
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
	securityGroups, err := s.list()
	if err != nil {
		return microerror.MaskAny(err)
	}

	for _, securityGroup := range securityGroups {
		_, err := s.Clients.EC2.DeleteSecurityGroup(&ec2.DeleteSecurityGroupInput{
			GroupId: securityGroup.GroupId,
		})
		if err != nil {
			return microerror.MaskAny(err)
		}
	}

	return nil
}

func (s SecurityGroup) ID() string {
	return s.id
}
