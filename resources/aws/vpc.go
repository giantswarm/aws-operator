package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	microerror "github.com/giantswarm/microkit/error"
)

type VPC struct {
	CidrBlock string
	Name      string
	id        string
	AWSEntity
}

func (v VPC) list() ([]*ec2.Vpc, error) {
	out, err := v.Clients.EC2.DescribeVpcs(&ec2.DescribeVpcsInput{
		Filters: []*ec2.Filter{
			&ec2.Filter{
				Name: aws.String(fmt.Sprintf("tag:%s", tagKeyName)),
				Values: []*string{
					aws.String(v.Name),
				},
			},
		},
	})
	return out.Vpcs, microerror.MaskAny(err)
}

func (v *VPC) checkIfExists() (bool, error) {
	vpcs, err := v.list()
	if err == nil && len(vpcs) > 0 {
		v.id = *vpcs[0].VpcId
	}
	return len(vpcs) > 0, microerror.MaskAny(err)
}

func (v *VPC) CreateIfNotExists() (bool, error) {
	exists, err := v.checkIfExists()
	if err != nil {
		return false, microerror.MaskAny(err)
	}

	if exists {
		return false, nil
	}

	if err := v.CreateOrFail(); err != nil {
		return false, microerror.MaskAny(err)
	}

	return true, nil
}

func (v *VPC) CreateOrFail() error {
	vpc, err := v.Clients.EC2.CreateVpc(&ec2.CreateVpcInput{
		CidrBlock: aws.String(v.CidrBlock),
	})
	if err != nil {
		return microerror.MaskAny(err)
	}
	vpcID := *vpc.Vpc.VpcId

	if err := v.Clients.EC2.WaitUntilVpcAvailable(&ec2.DescribeVpcsInput{
		VpcIds: []*string{
			aws.String(vpcID),
		},
	}); err != nil {
		return microerror.MaskAny(err)
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
		},
	}); err != nil {
		return microerror.MaskAny(err)
	}

	v.id = vpcID

	return nil
}

func (v *VPC) Delete() error {
	vpcs, err := v.list()
	if err != nil {
		return microerror.MaskAny(err)
	}

	for _, vpc := range vpcs {
		_, err := v.Clients.EC2.DeleteVpc(&ec2.DeleteVpcInput{
			VpcId: vpc.VpcId,
		})
		if err != nil {
			return microerror.MaskAny(err)
		}
	}

	return nil
}

func (v VPC) ID() string {
	return v.id
}
