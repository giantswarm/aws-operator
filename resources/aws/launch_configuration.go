package aws

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/giantswarm/microerror"
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
	defaultEBSVolumeMountPoint = "/dev/xvdh"
	// defaultEBSVolumeSize is expressed in GB.
	defaultEBSVolumeSize = 50
	defaultEBSVolumeType = "gp2"
	defaultEBSEncrypted  = true
)

// CreateIfNotExists creates the launch config if it doesn't exist.
func (lc *LaunchConfiguration) CreateIfNotExists() (bool, error) {
	if lc.Client == nil {
		return false, microerror.Mask(clientNotInitializedError)
	}

	exists, err := lc.checkIfExists()
	if err != nil {
		return false, microerror.Mask(err)
	}

	if exists {
		return false, nil
	}

	if err := lc.CreateOrFail(); err != nil {
		return false, microerror.Mask(err)
	}

	return true, nil
}

// CreateOrFail creates the launch config or returns the error.
func (lc *LaunchConfiguration) CreateOrFail() error {
	if lc.Client == nil {
		return microerror.Mask(clientNotInitializedError)
	}

	var ebsMount *autoscaling.BlockDeviceMapping
	if lc.EBSStorage {
		ebsMount = &autoscaling.BlockDeviceMapping{
			DeviceName: aws.String(defaultEBSVolumeMountPoint),
			Ebs: &autoscaling.Ebs{
				DeleteOnTermination: aws.Bool(true),
				VolumeSize:          aws.Int64(defaultEBSVolumeSize),
				VolumeType:          aws.String(defaultEBSVolumeType),
				Encrypted:           aws.Bool(defaultEBSEncrypted),
			},
		}
	}

	lcInput := &autoscaling.CreateLaunchConfigurationInput{
		LaunchConfigurationName: aws.String(lc.Name),
		IamInstanceProfile:      aws.String(lc.IamInstanceProfileName),
		ImageId:                 aws.String(lc.ImageID),
		InstanceType:            aws.String(lc.InstanceType),
		SecurityGroups: []*string{
			aws.String(lc.SecurityGroupID),
		},
		UserData:                 aws.String(lc.SmallCloudConfig),
		AssociatePublicIpAddress: aws.Bool(lc.AssociatePublicIpAddress),
		BlockDeviceMappings:      []*autoscaling.BlockDeviceMapping{ebsMount},
	}
	if lc.KeyName != "" {
		lcInput.KeyName = aws.String(lc.KeyName)
	}

	if _, err := lc.Client.CreateLaunchConfiguration(lcInput); err != nil {
		return microerror.Mask(err)
	}

	return nil
}

// Delete deletes the launch config.
func (lc *LaunchConfiguration) Delete() error {
	if lc.Client == nil {
		return microerror.Mask(clientNotInitializedError)
	}

	lcInput := &autoscaling.DeleteLaunchConfigurationInput{
		LaunchConfigurationName: aws.String(lc.Name),
	}
	if _, err := lc.Client.DeleteLaunchConfiguration(lcInput); err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func (lc *LaunchConfiguration) checkIfExists() (bool, error) {
	_, err := lc.findExisting()
	if IsNotFound(err) {
		return false, nil
	} else if err != nil {
		return false, microerror.Mask(err)
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
		return nil, microerror.Mask(err)
	}

	if len(launchConfigs.LaunchConfigurations) < 1 {
		return nil, microerror.Maskf(notFoundError, notFoundErrorFormat, LaunchConfigurationType, lc.Name)
	} else if len(launchConfigs.LaunchConfigurations) > 1 {
		return nil, microerror.Mask(tooManyResultsError)
	}

	return launchConfigs.LaunchConfigurations[0], nil
}
