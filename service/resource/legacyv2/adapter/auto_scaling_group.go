package adapter

import (
	"strconv"

	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/service/keyv2"
)

// template related to this adapter: service/templates/cloudformation/auto_scaling_group.yaml

type autoScalingGroupAdapter struct {
	ASGMaxSize             int
	ASGMinSize             int
	HealthCheckGracePeriod int
	MaxBatchSize           string
	MinInstancesInService  string
	RollingUpdatePauseTime string
	WorkerAZ               string
	WorkerSubnetID         string
}

func (a *autoScalingGroupAdapter) getAutoScalingGroup(customObject v1alpha1.AWSConfig, clients Clients) error {
	a.WorkerAZ = customObject.Spec.AWS.AZ
	workers := keyv2.WorkerCount(customObject)
	a.ASGMaxSize = workers
	a.ASGMinSize = workers
	a.MaxBatchSize = strconv.FormatFloat(asgMaxBatchSizeRatio, 'f', -1, 32)
	a.MinInstancesInService = strconv.FormatFloat(asgMinInstancesRatio, 'f', -1, 32)
	a.HealthCheckGracePeriod = gracePeriodSeconds
	a.RollingUpdatePauseTime = rollingUpdatePauseTime

	// subnet ID
	// TODO: remove this code once the subnet is created by cloudformation and add a
	// reference in the template
	subnetName := keyv2.SubnetName(customObject, suffixPrivate)
	subnetID, err := SubnetID(clients, subnetName)
	if err != nil {
		return microerror.Mask(err)
	}
	a.WorkerSubnetID = subnetID

	return nil
}
