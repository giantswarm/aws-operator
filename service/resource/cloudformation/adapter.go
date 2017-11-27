package cloudformation

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/giantswarm/awstpr"
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/service/key"
)

type hydrater func(awstpr.CustomObject, Clients) error

type adapter struct {
	ASGType string

	launchConfigAdapter
	autoScalingGroupAdapter
}

type launchConfigAdapter struct {
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
	HealthCheckGracePeriod int
	LoadBalancerName       string
	MaxBatchSize           string
	MinInstancesInService  string
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

func (l *launchConfigAdapter) getLaunchConfiguration(customObject awstpr.CustomObject, clients Clients) error {
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
	// TODO: remove this code once the security group is created by cloudformation
	// and add a reference in the template
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
	accountID := strings.Split(userArn, ":")[accountIDIndex]
	if err := ValidateAccountID(accountID); err != nil {
		return microerror.Mask(err)
	}

	clusterID := key.ClusterID(customObject)
	s3URI := fmt.Sprintf("%s-g8s-%s", accountID, clusterID)

	cloudConfigConfig := SmallCloudconfigConfig{
		MachineType: prefixWorker,
		Region:      customObject.Spec.AWS.Region,
		S3URI:       s3URI,
	}
	smallCloudConfig, err := SmallCloudconfig(cloudConfigConfig)
	if err != nil {
		return microerror.Mask(err)
	}
	l.SmallCloudConfig = smallCloudConfig

	return nil
}

func (a *autoScalingGroupAdapter) getAutoScalingGroup(customObject awstpr.CustomObject, clients Clients) error {
	a.AZ = customObject.Spec.AWS.AZ
	workers := key.WorkerCount(customObject)
	a.ASGMaxSize = workers
	a.ASGMinSize = workers
	a.MaxBatchSize = strconv.FormatFloat(asgMaxBatchSizeRatio, 'f', -1, 32)
	a.MinInstancesInService = strconv.FormatFloat(asgMinInstancesRatio, 'f', -1, 32)
	a.HealthCheckGracePeriod = gracePeriodSeconds
	a.RollingUpdatePauseTime = rollingUpdatePauseTime

	// load balancer name
	// TODO: remove this code once the ingress load balancer is created by cloudformation
	// and add a reference in the template
	lbName, err := ingressLoadBalancerName(customObject)
	if err != nil {
		return microerror.Mask(err)
	}
	a.LoadBalancerName = lbName

	// subnet ID
	// TODO: remove this code once the subnet is created by cloudformation and add a
	// reference in the template
	subnetName := key.SubnetName(customObject, suffixPublic)
	describeSubnetInput := &ec2.DescribeSubnetsInput{
		Filters: []*ec2.Filter{
			{
				Name: aws.String(fmt.Sprintf("tag:%s", tagKeyName)),
				Values: []*string{
					aws.String(subnetName),
				},
			},
		},
	}
	output, err := clients.EC2.DescribeSubnets(describeSubnetInput)
	if err != nil {
		return microerror.Mask(err)
	}
	if len(output.Subnets) > 1 {
		return microerror.Mask(tooManyResultsError)
	}

	a.SubnetID = *output.Subnets[0].SubnetId

	return nil
}
