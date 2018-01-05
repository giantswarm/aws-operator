package adapter

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/service/keyv2"
)

// template related to this adapter: service/templates/cloudformation/instance.yaml

type instanceAdapter struct {
	MasterAZ               string
	MasterImageID          string
	MasterInstanceType     string
	MasterSecurityGroupID  string
	MasterSmallCloudConfig string
}

func (i *instanceAdapter) getInstance(customObject v1alpha1.AWSConfig, clients Clients) error {
	if len(customObject.Spec.AWS.Masters) == 0 {
		return microerror.Mask(invalidConfigError)
	}

	i.MasterAZ = keyv2.AvailabilityZone(customObject)
	i.MasterImageID = keyv2.MasterImageID(customObject)
	i.MasterInstanceType = keyv2.MasterInstanceType(customObject)

	// security group
	// TODO: remove this code once the security group is created by cloudformation
	// and add a reference in the template
	groupName := keyv2.SecurityGroupName(customObject, prefixMaster)
	describeSgInput := &ec2.DescribeSecurityGroupsInput{
		Filters: []*ec2.Filter{
			{
				Name: aws.String(subnetDescription),
				Values: []*string{
					aws.String(groupName),
				},
			},
			{
				Name: aws.String(subnetGroupName),
				Values: []*string{
					aws.String(groupName),
				},
			},
		},
	}
	output, err := clients.EC2.DescribeSecurityGroups(describeSgInput)
	if err != nil {
		return microerror.Mask(err)
	}
	if len(output.SecurityGroups) > 1 {
		return microerror.Mask(tooManyResultsError)
	}
	i.MasterSecurityGroupID = *output.SecurityGroups[0].GroupId

	accountID, err := AccountID(clients)
	if err != nil {
		return microerror.Mask(err)
	}

	clusterID := keyv2.ClusterID(customObject)
	s3URI := fmt.Sprintf("%s-g8s-%s", accountID, clusterID)

	cloudConfigConfig := SmallCloudconfigConfig{
		MachineType:    prefixMaster,
		Region:         customObject.Spec.AWS.Region,
		S3URI:          s3URI,
		ClusterVersion: keyv2.ClusterVersion(customObject),
	}
	smallCloudConfig, err := SmallCloudconfig(cloudConfigConfig)
	if err != nil {
		return microerror.Mask(err)
	}
	i.MasterSmallCloudConfig = smallCloudConfig

	return nil
}
