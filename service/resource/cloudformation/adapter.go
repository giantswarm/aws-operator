package cloudformation

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/giantswarm/awstpr"
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/service/key"
	"github.com/giantswarm/aws-operator/service/resource/legacy"
)

type hydrater func(awstpr.CustomObject, Clients) error

type adapter struct {
	ASGType string

	lauchConfigAdapter
	autoScalingGroupAdapter
}

type lauchConfigAdapter struct {
	AssociatePublicIPAddress bool
	BlockDeviceMappings      []BlockDeviceMapping
	IAMInstanceProfileName   string
	ImageID                  string
	InstanceType             string
	SecurityGroupID          string
	SmallCloudConfig         string
}

type BlockDeviceMapping struct {
	DeleteOnTermination bool
	DeviceName          string
	VolumeSize          int
	VolumeType          string
}

type autoScalingGroupAdapter struct {
	ASGMaxSize             int
	ASGMinSize             int
	AZ                     string
	HealthCheckGracePeriod string
	LoadBalancerName       string
	MaxBatchSize           string
	MinInstancesInService  int
	RollingUpdatePauseTime string
	SubnetID               string
}

func newAdapter(customObject awstpr.CustomObject, clients Clients) (adapter, error) {
	a := adapter{}

	a.ASGType = prefixWorker

	hydraters := []hydrater{
		a.getAutoScalingGroup,
		a.getLaunchConfiguration,
	}

	for _, h := range hydraters {
		if err := h(customObject, clients); err != nil {
			return adapter{}, microerror.Mask(err)
		}
	}

	return a, nil
}

func (l *lauchConfigAdapter) getLaunchConfiguration(customObject awstpr.CustomObject, clients Clients) error {
	if len(customObject.Spec.AWS.Workers) == 0 {
		return microerror.Mask(invalidConfigError)
	}

	l.ImageID = customObject.Spec.AWS.Workers[0].ImageID
	l.InstanceType = customObject.Spec.AWS.Workers[0].InstanceType
	l.IAMInstanceProfileName = key.InstanceProfileName(customObject, prefixWorker)
	l.AssociatePublicIPAddress = true

	l.BlockDeviceMappings = []BlockDeviceMapping{
		BlockDeviceMapping{
			DeleteOnTermination: true,
			DeviceName:          defaultEBSVolumeMountPoint,
			VolumeSize:          defaultEBSVolumeSize,
			VolumeType:          defaultEBSVolumeType,
		},
	}

	// security group
	groupName := key.SecurityGroupName(customObject, prefixWorker)
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
	l.SecurityGroupID = *output.SecurityGroups[0].GroupId

	// cloud config
	resp, err := clients.IAM.GetUser(&iam.GetUserInput{})
	if err != nil {
		return microerror.Mask(err)
	}
	userArn := *resp.User.Arn
	accountID := strings.Split(userArn, ":")[4]
	clusterID := key.ClusterID(customObject)
	s3URI := fmt.Sprintf("%s-g8s-%s", accountID, clusterID)

	cloudConfigConfig := legacy.SmallCloudconfigConfig{
		MachineType: prefixWorker,
		Region:      customObject.Spec.AWS.Region,
		S3URI:       s3URI,
	}
	smallCloudConfig, err := legacy.SmallCloudconfig(cloudConfigConfig)
	if err != nil {
		return microerror.Mask(err)
	}
	l.SmallCloudConfig = smallCloudConfig

	return nil
}

func (a *autoScalingGroupAdapter) getAutoScalingGroup(customObject awstpr.CustomObject, clients Clients) error {
	a.AZ = customObject.Spec.AWS.AZ

	return nil
}
