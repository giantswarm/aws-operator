package aws

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	microerror "github.com/giantswarm/microkit/error"
)

// LaunchConfiguration is a template for launching EC2 instances into an auto
// scaling group.
type LaunchConfiguration struct {
	AssociatePublicIpAddress bool
	EBSStorage               bool
	IamInstanceProfileName   string
	ImageID                  string
	InstanceType             string
	KeyName                  string
	Name                     string
	SecurityGroupID          string
	SmallCloudConfig         string

	// Dependencies
	Client *autoscaling.AutoScaling
}

const (
	defaultMountPoint = "/dev/xvdh"
	// defaultVolumeSize is expressed in GB.
	defaultVolumeSize = 50
	defaultVolumeType = "gp2"
)

// CreateIfNotExists creates the launch config if it doesn't exist.
func (lc *LaunchConfiguration) CreateIfNotExists() (bool, error) {
	if lc.Client == nil {
		return false, microerror.MaskAny(clientNotInitializedError)
	}

	exists, err := lc.checkIfExists()
	if err != nil {
		return false, microerror.MaskAny(err)
	}

	if exists {
		return false, nil
	}

	if err := lc.CreateOrFail(); err != nil {
		return false, microerror.MaskAny(err)
	}

	return true, nil
}

// CreateOrFail creates the launch config or returns the error.
func (lc *LaunchConfiguration) CreateOrFail() error {
	if lc.Client == nil {
		return microerror.MaskAny(clientNotInitializedError)
	}

	var ebsMount *autoscaling.BlockDeviceMapping
	if lc.EBSStorage {
		ebsMount = &autoscaling.BlockDeviceMapping{
			DeviceName: aws.String(defaultMountPoint),
			Ebs: &autoscaling.Ebs{
				DeleteOnTermination: aws.Bool(true),
				VolumeSize:          aws.Int64(defaultVolumeSize),
				VolumeType:          aws.String(defaultVolumeType),
			},
		}
	}

	lcInput := &autoscaling.CreateLaunchConfigurationInput{
		LaunchConfigurationName: aws.String(lc.Name),
		IamInstanceProfile:      aws.String(lc.IamInstanceProfileName),
		ImageId:                 aws.String(lc.ImageID),
		InstanceType:            aws.String(lc.InstanceType),
		KeyName:                 aws.String(lc.KeyName),
		SecurityGroups: []*string{
			aws.String(lc.SecurityGroupID),
		},
		UserData:                 aws.String(lc.SmallCloudConfig),
		AssociatePublicIpAddress: aws.Bool(lc.AssociatePublicIpAddress),
		BlockDeviceMappings:      []*autoscaling.BlockDeviceMapping{ebsMount},
	}

	if _, err := lc.Client.CreateLaunchConfiguration(lcInput); err != nil {
		return microerror.MaskAny(err)
	}

	return nil
}

// Delete deletes the launch config.
func (lc *LaunchConfiguration) Delete() error {
	if lc.Client == nil {
		return microerror.MaskAny(clientNotInitializedError)
	}

	lcInput := &autoscaling.DeleteLaunchConfigurationInput{
		LaunchConfigurationName: aws.String(lc.Name),
	}
	if _, err := lc.Client.DeleteLaunchConfiguration(lcInput); err != nil {
		return microerror.MaskAny(err)
	}

	return nil
}

func (lc *LaunchConfiguration) checkIfExists() (bool, error) {
	_, err := lc.findExisting()
	if IsNotFound(err) {
		return false, nil
	} else if err != nil {
		return false, microerror.MaskAny(err)
	}

	return true, nil
}

func (lc *LaunchConfiguration) findExisting() (*autoscaling.LaunchConfiguration, error) {
	launchConfigs, err := lc.Client.DescribeLaunchConfigurations(&autoscaling.DescribeLaunchConfigurationsInput{
		LaunchConfigurationNames: []*string{
			aws.String(lc.Name),
		},
	})
	if err != nil {
		return nil, microerror.MaskAny(err)
	}

	if len(launchConfigs.LaunchConfigurations) < 1 {
		return nil, microerror.MaskAnyf(notFoundError, notFoundErrorFormat, LaunchConfigurationType, lc.Name)
	} else if len(launchConfigs.LaunchConfigurations) > 1 {
		return nil, microerror.MaskAny(tooManyResultsError)
	}

	return launchConfigs.LaunchConfigurations[0], nil
}
