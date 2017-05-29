package aws

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/cenkalti/backoff"
	microerror "github.com/giantswarm/microkit/error"
	micrologger "github.com/giantswarm/microkit/logger"

	awsclient "github.com/giantswarm/aws-operator/client/aws"
)

type Subnet struct {
	AvailabilityZone string
	CidrBlock        string
	Name             string
	VpcID            string
	id               string
	// Dependencies.
	Logger micrologger.Logger
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
		return nil, microerror.MaskAnyf(notFoundError, notFoundErrorFormat, SubnetType, s.Name)
	} else if len(subnets.Subnets) > 1 {
		return nil, microerror.MaskAny(tooManyResultsError)
	}

	return subnets.Subnets[0], nil
}

func (s *Subnet) checkIfExists() (bool, error) {
	_, err := s.findExisting()
	if IsNotFound(err) {
		return false, nil
	} else if err != nil {
		return false, microerror.MaskAny(err)
	}

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
	subnetID, err := s.GetID()
	if err != nil {
		return microerror.MaskAny(err)
	}

	deleteOperation := func() error {
		if _, err := s.Clients.EC2.DeleteSubnet(&ec2.DeleteSubnetInput{
			SubnetId: aws.String(subnetID),
		}); err != nil {
			return microerror.MaskAny(err)
		}
		return nil
	}
	deleteNotify := NewNotify(s.Logger, "deleting subnet")
	if err := backoff.RetryNotify(deleteOperation, NewCustomExponentialBackoff(), deleteNotify); err != nil {
		return microerror.MaskAny(err)
	}

	return nil
}

func (s Subnet) GetID() (string, error) {
	if s.id != "" {
		return s.id, nil
	}

	subnet, err := s.findExisting()
	if err != nil {
		return "", microerror.MaskAny(err)
	}

	return *subnet.SubnetId, nil
}

func (s *Subnet) MakePublic(routeTable *RouteTable) error {
	routeTableID, err := routeTable.GetID()
	if err != nil {
		return microerror.MaskAny(err)
	}

	subnetID, err := s.GetID()
	if err != nil {
		return microerror.MaskAny(err)
	}

	if _, err := s.Clients.EC2.AssociateRouteTable(&ec2.AssociateRouteTableInput{
		RouteTableId: aws.String(routeTableID),
		SubnetId:     aws.String(subnetID),
	}); err != nil {
		if !strings.Contains(err.Error(), awsclient.AlreadyAssociated) {
			return microerror.MaskAny(err)
		}
	}

	if _, err := s.Clients.EC2.ModifySubnetAttribute(&ec2.ModifySubnetAttributeInput{
		MapPublicIpOnLaunch: &ec2.AttributeBooleanValue{
			Value: aws.Bool(true),
		},
		SubnetId: aws.String(subnetID),
	}); err != nil {
		return microerror.MaskAny(err)
	}

	return nil
}
