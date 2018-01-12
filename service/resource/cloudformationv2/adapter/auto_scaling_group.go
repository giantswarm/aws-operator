package adapter

import (
	"strconv"

	"github.com/giantswarm/aws-operator/service/keyv2"
)

// template related to this adapter: service/templates/cloudformation/guest/auto_scaling_group.yaml

type autoScalingGroupAdapter struct {
	ASGMaxSize             int
	ASGMinSize             int
	HealthCheckGracePeriod int
	MaxBatchSize           string
	MinInstancesInService  string
	RollingUpdatePauseTime string
	WorkerAZ               string
}

func (a *autoScalingGroupAdapter) getAutoScalingGroup(cfg Config) error {
	a.WorkerAZ = cfg.CustomObject.Spec.AWS.AZ
	workers := keyv2.WorkerCount(cfg.CustomObject)
	a.ASGMaxSize = workers
	a.ASGMinSize = workers
	a.MaxBatchSize = strconv.FormatFloat(asgMaxBatchSizeRatio, 'f', -1, 32)
	a.MinInstancesInService = strconv.FormatFloat(asgMinInstancesRatio, 'f', -1, 32)
	a.HealthCheckGracePeriod = gracePeriodSeconds
	a.RollingUpdatePauseTime = rollingUpdatePauseTime

	return nil
}
