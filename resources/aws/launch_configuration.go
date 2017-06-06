package aws

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	microerror "github.com/giantswarm/microkit/error"
)

// LaunchConfiguration is a template for launching EC2 instances into an auto
// scaling group.
type LaunchConfiguration struct {
	Name                   string
	IamInstanceProfileName string
	ImageID                string
	InstanceType           string
	KeyName                string
	SecurityGroupID        string
	SmallCloudConfig       string

	// Dependencies
	Client *autoscaling.AutoScaling
}

// CreateIfNotExists creates the launch config if it doesn't exist.
func (lc *LaunchConfiguration) CreateIfNotExists() (bool, error) {
	if lc.Client == nil {
		return false, microerror.MaskAny(clientNotInitializedError)
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

	if _, err := lc.Client.CreateLaunchConfiguration(&autoscaling.CreateLaunchConfigurationInput{
		LaunchConfigurationName: aws.String(lc.Name),
		IamInstanceProfile:      aws.String(lc.IamInstanceProfileName),
		ImageId:                 aws.String(lc.ImageID),
		InstanceType:            aws.String(lc.InstanceType),
		KeyName:                 aws.String(lc.KeyName),
		SecurityGroups: []*string{
			aws.String(lc.SecurityGroupID),
		},
		UserData: aws.String(lc.SmallCloudConfig),
	}); err != nil {
		return microerror.MaskAny(err)
	}

	return nil
}

// Delete deletes the launch config.
func (lc *LaunchConfiguration) Delete() error {
	if lc.Client == nil {
		return microerror.MaskAny(clientNotInitializedError)
	}

	if _, err := lc.Client.DeleteLaunchConfiguration(&autoscaling.DeleteLaunchConfigurationInput{
		LaunchConfigurationName: aws.String(lc.Name),
	}); err != nil {
		return microerror.MaskAny(err)
	}

	return nil
}
