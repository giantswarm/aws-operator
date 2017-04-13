package aws

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	microerror "github.com/giantswarm/microkit/error"

	awsclient "github.com/giantswarm/aws-operator/client/aws"
)

type Subnet struct {
	AvailabilityZone string
	CidrBlock        string
	Name             string
	VpcID            string
	id               string
	AWSEntity
}

func (s Subnet) findExisting() (*ec2.Subnet, error) {
	subnets, err := s.Clients.EC2.DescribeSubnets(&ec2.DescribeSubnetsInput{
		Filters: []*ec2.Filter{
			&ec2.Filter{
				Name: aws.String(fmt.Sprintf("tag:%s", tagKeyName)),
				Values: []*string{
					aws.String(s.Name),
				},
			},
		},
	})
	if err != nil {
		return nil, microerror.MaskAny(err)
	}

	if len(subnets.Subnets) < 1 {
		return nil, microerror.MaskAny(subnetFindError)
	}

	return subnets.Subnets[0], nil
}

func (s *Subnet) checkIfExists() (bool, error) {
	subnet, err := s.findExisting()
	if err != nil {
		if strings.Contains(err.Error(), subnetFindError.Error()) {
			return false, nil
		}
		return false, microerror.MaskAny(err)
	}

	s.id = *subnet.SubnetId
	return true, nil
}

func (s *Subnet) CreateIfNotExists() (bool, error) {
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

func (s *Subnet) CreateOrFail() error {
	subnet, err := s.Clients.EC2.CreateSubnet(&ec2.CreateSubnetInput{
		AvailabilityZone: aws.String(s.AvailabilityZone),
		CidrBlock:        aws.String(s.CidrBlock),
		VpcId:            aws.String(s.VpcID),
	})
	if err != nil {
		return microerror.MaskAny(err)
	}

	if _, err := s.Clients.EC2.CreateTags(&ec2.CreateTagsInput{
		Resources: []*string{subnet.Subnet.SubnetId},
		Tags: []*ec2.Tag{
			{
				Key:   aws.String(tagKeyName),
				Value: aws.String(s.Name),
			},
		},
	}); err != nil {
		return microerror.MaskAny(err)
	}

	s.id = *subnet.Subnet.SubnetId

	return nil
}

func (s *Subnet) Delete() error {
	subnet, err := s.findExisting()
	if err != nil {
		return microerror.MaskAny(err)
	}

	if _, err := s.Clients.EC2.DeleteSubnet(&ec2.DeleteSubnetInput{
		SubnetId: subnet.SubnetId,
	}); err != nil {
		return microerror.MaskAny(err)
	}

	return nil
}

func (s Subnet) ID() string {
	return s.id
}

func (s *Subnet) MakePublic(routeTable *RouteTable) error {
	if _, err := s.Clients.EC2.AssociateRouteTable(&ec2.AssociateRouteTableInput{
		RouteTableId: aws.String(routeTable.ID()),
		SubnetId:     aws.String(s.ID()),
	}); err != nil {
		if !strings.Contains(err.Error(), awsclient.AlreadyAssociated) {
			return microerror.MaskAny(err)
		}
	}

	if _, err := s.Clients.EC2.ModifySubnetAttribute(&ec2.ModifySubnetAttributeInput{
		MapPublicIpOnLaunch: &ec2.AttributeBooleanValue{
			Value: aws.Bool(true),
		},
		SubnetId: aws.String(s.ID()),
	}); err != nil {
		return microerror.MaskAny(err)
	}

	return nil
}
