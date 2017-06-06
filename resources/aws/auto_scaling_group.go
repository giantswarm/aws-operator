package aws

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	microerror "github.com/giantswarm/microkit/error"
)

type AutoScalingGroup struct {
	Client                  *autoscaling.AutoScaling
	Name                    string
	MinSize                 int
	MaxSize                 int
	AvailabilityZone        string
	LaunchConfigurationName string
	LoadBalancerName        string
	VPCZoneIdentifier       string
	HealthCheckGracePeriod  int
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
		ForceDelete:          aws.Bool(true),
	}

	if _, err := asg.Client.DeleteAutoScalingGroup(params); err != nil {
		return microerror.MaskAny(err)
	}

	return nil
}
