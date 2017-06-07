package aws

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	microerror "github.com/giantswarm/microkit/error"
)

type AutoScalingGroup struct {
	// AvailabilityZone is the AZ the instances will be placed in.
	AvailabilityZone string
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

func (asg *AutoScalingGroup) CreateOrFail() error {
	if asg.Client == nil {
		return microerror.MaskAny(clientNotInitializedError)
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
				Value:             aws.String(asg.Name),
			},
		},
	}

	if _, err := asg.Client.CreateAutoScalingGroup(params); err != nil {
		return microerror.MaskAny(err)
	}

	return nil
}

func (asg *AutoScalingGroup) Delete() error {
	if asg.Client == nil {
		return microerror.MaskAny(clientNotInitializedError)
	}

	params := &autoscaling.DeleteAutoScalingGroupInput{
		AutoScalingGroupName: aws.String(asg.Name),
		// Delete instances too.
		ForceDelete: aws.Bool(true),
	}

	if _, err := asg.Client.DeleteAutoScalingGroup(params); err != nil {
		return microerror.MaskAny(err)
	}

	return nil
}
