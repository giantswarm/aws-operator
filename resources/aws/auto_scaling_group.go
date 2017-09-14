package aws

import (
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/giantswarm/microerror"
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

// FindLegacy returns true if there is an ASG created for the same cluster name
// but not being part of a CloudFormation stack.
func (asg *AutoScalingGroup) FindLegacy() (bool, error) {
	a, err := asg.findExisting()
	if IsNotFound(err) {
		return false, nil
	} else if err != nil {
		return false, microerror.Mask(err)
	}

	return isLegacyASG(asg.Name, asg.ClusterID, a.Tags), nil
}

// DetachInstances detaches instances from this ASG and returns their details
func (asg *AutoScalingGroup) DetachInstances() ([]*autoscaling.Instance, error) {
	autoScalingGroup, err := asg.findExisting()
	if err != nil {
		return []*autoscaling.Instance{}, microerror.Mask(err)
	}

	for _, i := range autoScalingGroup.Instances {
		_, err = asg.Client.DetachInstances(&autoscaling.DetachInstancesInput{
			AutoScalingGroupName:           aws.String(asg.Name),
			InstanceIds:                    []*string{i.InstanceId},
			ShouldDecrementDesiredCapacity: aws.Bool(false),
		})
		if err != nil {
			return []*autoscaling.Instance{}, microerror.Mask(err)
		}
	}

	return autoScalingGroup.Instances, nil
}

// AttachInstances returns details of the instances attached to this ASG
func (asg *AutoScalingGroup) AttachInstances(instances []*autoscaling.Instance) error {
	// save current termination policies
	autoScalingGroup, err := asg.findExisting()
	if err != nil {
		return microerror.Mask(err)
	}
	backTerminationPolicies := autoScalingGroup.TerminationPolicies

	// set termination policy to newest first so that the old instances don't get removed
	terminationPolicy := newestFirstTerminationPolicy
	params := &autoscaling.UpdateAutoScalingGroupInput{
		AutoScalingGroupName: aws.String(asg.Name),
		TerminationPolicies:  []*string{&terminationPolicy},
	}
	if _, err := asg.Client.UpdateAutoScalingGroup(params); err != nil {
		return microerror.Mask(err)
	}

	var ids []*string
	for _, i := range instances {
		ids = append(ids, i.InstanceId)
	}

	_, err = asg.Client.AttachInstances(&autoscaling.AttachInstancesInput{
		AutoScalingGroupName: aws.String(asg.Name),
		InstanceIds:          ids,
	})
	if err != nil {
		return microerror.Mask(err)
	}

	// back to original termination policies
	params = &autoscaling.UpdateAutoScalingGroupInput{
		AutoScalingGroupName: aws.String(asg.Name),
		TerminationPolicies:  backTerminationPolicies,
	}
	if _, err := asg.Client.UpdateAutoScalingGroup(params); err != nil {
		return microerror.Mask(err)
	}
	return nil
}

func isLegacyASG(name, clusterID string, tags []*autoscaling.TagDescription) bool {
	var keyNameFound, keyClusterFound bool
	for _, td := range tags {
		if *td.Key == tagKeyName && *td.Value == name {
			keyNameFound = true
		} else if *td.Key == tagKeyCluster && *td.Value == clusterID {
			keyClusterFound = true
		} else if strings.HasPrefix(*td.Key, "aws:cloudformation:") {
			return false
		}
	}

	return keyNameFound && keyClusterFound
}
