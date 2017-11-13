package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/cenkalti/backoff"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
)

type VPC struct {
	CidrBlock        string
	InstallationName string
	Name             string
	id               string
	// Dependencies.
	Logger micrologger.Logger
	AWSEntity
}

func (v VPC) findExisting() (*ec2.Vpc, error) {
	vpcs, err := v.Clients.EC2.DescribeVpcs(&ec2.DescribeVpcsInput{
		Filters: []*ec2.Filter{
			{
				Name: aws.String(fmt.Sprintf("tag:%s", tagKeyName)),
				Values: []*string{
					aws.String(v.Name),
				},
			},
		},
	})
	if err != nil {
		return nil, microerror.Mask(err)
	}

	if len(vpcs.Vpcs) < 1 {
		return nil, microerror.Maskf(notFoundError, notFoundErrorFormat, VPCType, v.Name)
	} else if len(vpcs.Vpcs) > 1 {
		return nil, microerror.Mask(tooManyResultsError)
	}

	return vpcs.Vpcs[0], nil
}

func (v *VPC) checkIfExists() (bool, error) {
	_, err := v.findExisting()
	if IsNotFound(err) {
		return false, nil
	} else if err != nil {
		return false, microerror.Mask(err)
	}

	return true, nil
}

func (v *VPC) CreateIfNotExists() (bool, error) {
	exists, err := v.checkIfExists()
	if err != nil {
		return false, microerror.Mask(err)
	}

	if exists {
		return false, nil
	}

	if err := v.CreateOrFail(); err != nil {
		return false, microerror.Mask(err)
	}

	return true, nil
}

func (v *VPC) CreateOrFail() error {
	vpc, err := v.Clients.EC2.CreateVpc(&ec2.CreateVpcInput{
		CidrBlock: aws.String(v.CidrBlock),
	})
	if err != nil {
		return microerror.Mask(err)
	}
	vpcID := *vpc.Vpc.VpcId

	if err := v.Clients.EC2.WaitUntilVpcAvailable(&ec2.DescribeVpcsInput{
		VpcIds: []*string{
			aws.String(vpcID),
		},
	}); err != nil {
		return microerror.Mask(err)
	}

	if _, err := v.Clients.EC2.CreateTags(&ec2.CreateTagsInput{
		Resources: []*string{
			aws.String(vpcID),
		},
		Tags: []*ec2.Tag{
			{
				Key:   aws.String(tagKeyName),
				Value: aws.String(v.Name),
			},
			{
				Key:   aws.String(tagKeyInstallation),
				Value: aws.String(v.InstallationName),
			},
		},
	}); err != nil {
		return microerror.Mask(err)
	}

	// These attributes are required for a VPC with private Hosted Zones.
	if _, err := v.Clients.EC2.ModifyVpcAttribute(&ec2.ModifyVpcAttributeInput{
		EnableDnsHostnames: &ec2.AttributeBooleanValue{
			Value: aws.Bool(true),
		},
		VpcId: aws.String(vpcID),
	}); err != nil {
		return microerror.Mask(err)
	}

	if _, err := v.Clients.EC2.ModifyVpcAttribute(&ec2.ModifyVpcAttributeInput{
		EnableDnsSupport: &ec2.AttributeBooleanValue{
			Value: aws.Bool(true),
		},
		VpcId: aws.String(vpcID),
	}); err != nil {
		return microerror.Mask(err)
	}

	v.id = vpcID

	return nil
}

func (v *VPC) Delete() error {
	vpc, err := v.findExisting()
	if err != nil {
		return microerror.Mask(err)
	}

	deleteOperation := func() error {
		if _, err := v.Clients.EC2.DeleteVpc(&ec2.DeleteVpcInput{
			VpcId: vpc.VpcId,
		}); err != nil {
			return microerror.Mask(err)
		}
		return nil
	}
	deleteNotify := NewNotify(v.Logger, "deleting vpc")
	if err := backoff.RetryNotify(deleteOperation, NewCustomExponentialBackoff(), deleteNotify); err != nil {
		return microerror.Mask(err)
	}

	return nil
}

// GetID retrieves the ID from the API if it isn't defined on the VPC struct.
func (v VPC) GetID() (string, error) {
	if v.id != "" {
		return v.id, nil
	}

	vpc, err := v.findExisting()
	if err != nil {
		return "", microerror.Mask(err)
	}

	return *vpc.VpcId, nil
}

// List lists the VPCs for this installation.
func (v VPC) List() ([]VPC, error) {
	vpcs := []VPC{}

	resp, err := v.Clients.EC2.DescribeVpcs(&ec2.DescribeVpcsInput{
		Filters: []*ec2.Filter{
			{
				Name: aws.String(fmt.Sprintf("tag:%s", tagKeyInstallation)),
				Values: []*string{
					aws.String(v.InstallationName),
				},
			},
		},
	})
	if err != nil {
		return []VPC{}, microerror.Mask(err)
	}

	for _, vpc := range resp.Vpcs {
		for _, tag := range vpc.Tags {
			if *tag.Key == "Name" {
				vpcs = append(vpcs, VPC{Name: *tag.Value})
			}
		}
	}

	return vpcs, nil
}
