package aws

import (
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

func (s *SecurityGroup) CreateIfNotExists() (bool, error) {
	return false, microerror.MaskAny(notImplementedMethodError)
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
	if _, err := s.Clients.EC2.DeleteSecurityGroup(&ec2.DeleteSecurityGroupInput{
		GroupId: aws.String(s.ID()),
	}); err != nil {
		return microerror.MaskAny(err)
	}

	return nil
}

func (s *SecurityGroup) ID() string {
	return s.id
}
