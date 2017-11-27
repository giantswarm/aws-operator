package aws

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/giantswarm/microerror"
)

const (
	ASGMetricsGranularity = "1Minute"
)

type AutoScalingGroup struct {
	// AvailabilityZone is the AZ the instances will be placed in.
	AvailabilityZone string
	// ClusterID is the ID of the cluster.
	ClusterID string
	// HealthCheckGracePeriod is the time, in seconds, that the instances are
	// given after boot before the healthchecks start.
	HealthCheckGracePeriod int
	// LaunchConfigurationName is the name of the Launch Configuration used for the instances.
	LaunchConfigurationName string
	// LoadBalancerName is the name of the existing ELB that will be placed in
	// the ASG to front the instances.
	LoadBalancerName string
	// MaxSize is the maximum amount of instances that will be created in this ASG.
	MaxSize int
	// MinSize is the minimum amount of instances in this ASG. There will never be
	// less than MinSize instances running.
	MinSize int
	// Name is the ASG name.
	Name string
	// VPCZoneIdentifier is the Subnet ID of the subnet the instances should be
	// placed in.
	VPCZoneIdentifier string

	// Dependencies.
	Client *autoscaling.AutoScaling
}

const (
	AutoScalingGroupType resourceType = "auto scaling group"
)

func (asg *AutoScalingGroup) CreateIfNotExists() (bool, error) {
	if asg.Client == nil {
		return false, microerror.Mask(clientNotInitializedError)
	}

	exists, err := asg.checkIfExists()
	if err != nil {
		return false, microerror.Mask(err)
	}
	if exists {
		return false, nil
	}

	if err := asg.CreateOrFail(); err != nil {
		return false, microerror.Mask(err)
	}

	return true, nil
}

func (asg *AutoScalingGroup) CreateOrFail() error {
	if asg.Client == nil {
		return microerror.Mask(clientNotInitializedError)
	}

	params := &autoscaling.CreateAutoScalingGroupInput{
		AutoScalingGroupName: aws.String(asg.Name),
		MaxSize:              aws.Int64(int64(asg.MaxSize)),
		MinSize:              aws.Int64(int64(asg.MinSize)),
		AvailabilityZones: []*string{
			aws.String(asg.AvailabilityZone),
		},
		LaunchConfigurationName: aws.String(asg.LaunchConfigurationName),
		LoadBalancerNames: []*string{
			aws.String(asg.LoadBalancerName),
		},
		VPCZoneIdentifier:      aws.String(asg.VPCZoneIdentifier),
		HealthCheckGracePeriod: aws.Int64(int64(asg.HealthCheckGracePeriod)),
		Tags: []*autoscaling.Tag{
			{
				Key:               aws.String(tagKeyName),
				PropagateAtLaunch: aws.Bool(true),
				Value:             aws.String(asg.Name),
			},
			{
				Key:               aws.String(tagKeyCluster),
				PropagateAtLaunch: aws.Bool(true),
				Value:             aws.String(asg.ClusterID),
			},
		},
	}

	if _, err := asg.Client.CreateAutoScalingGroup(params); err != nil {
		return microerror.Mask(err)
	}

	if _, err := asg.Client.EnableMetricsCollection(&autoscaling.EnableMetricsCollectionInput{
		AutoScalingGroupName: aws.String(asg.Name),
		Granularity:          aws.String(ASGMetricsGranularity),
	}); err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func (asg *AutoScalingGroup) Delete() error {
	if asg.Client == nil {
		return microerror.Mask(clientNotInitializedError)
	}

	params := &autoscaling.DeleteAutoScalingGroupInput{
		AutoScalingGroupName: aws.String(asg.Name),
		// Delete instances too.
		ForceDelete: aws.Bool(true),
	}

	if _, err := asg.Client.DeleteAutoScalingGroup(params); err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func (asg *AutoScalingGroup) Update() error {
	if asg.Client == nil {
		return microerror.Mask(clientNotInitializedError)
	}

	params := &autoscaling.UpdateAutoScalingGroupInput{
		AutoScalingGroupName: aws.String(asg.Name),
		MinSize:              aws.Int64(int64(asg.MinSize)),
		MaxSize:              aws.Int64(int64(asg.MaxSize)),
	}

	if _, err := asg.Client.UpdateAutoScalingGroup(params); err != nil {
		return microerror.Mask(err)
	}
	return nil
}

func (asg *AutoScalingGroup) checkIfExists() (bool, error) {
	_, err := asg.findExisting()
	if IsNotFound(err) {
		return false, nil
	} else if err != nil {
		return false, microerror.Mask(err)
	}

	return true, nil
}

func (asg *AutoScalingGroup) findExisting() (*autoscaling.Group, error) {
	autoScalingGroups, err := asg.Client.DescribeAutoScalingGroups(&autoscaling.DescribeAutoScalingGroupsInput{
		AutoScalingGroupNames: []*string{
			aws.String(asg.Name),
		},
	})
	if err != nil {
		return nil, microerror.Mask(err)
	}

	if len(autoScalingGroups.AutoScalingGroups) < 1 {
		return nil, microerror.Maskf(notFoundError, notFoundErrorFormat, AutoScalingGroupType, asg.Name)
	} else if len(autoScalingGroups.AutoScalingGroups) > 1 {
		return nil, microerror.Mask(tooManyResultsError)
	}

	return autoScalingGroups.AutoScalingGroups[0], nil
}
