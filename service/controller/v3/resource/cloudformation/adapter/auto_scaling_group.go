package adapter

import (
	"strconv"

	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/service/controller/v3/key"
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
	workers := key.WorkerCount(cfg.CustomObject)
	if workers <= 0 {
		return microerror.Maskf(invalidConfigError, "at least 1 worker required, found %d", workers)
	}

	a.WorkerAZ = key.AvailabilityZone(cfg.CustomObject)
	a.ASGMaxSize = workers
	a.ASGMinSize = workers
	a.MaxBatchSize = workerCountRatio(workers, asgMaxBatchSizeRatio)
	a.MinInstancesInService = workerCountRatio(workers, asgMinInstancesRatio)
	a.HealthCheckGracePeriod = gracePeriodSeconds
	a.RollingUpdatePauseTime = rollingUpdatePauseTime

	return nil
}

func workerCountRatio(workers int, ratio float32) string {
	value := float32(workers) * ratio
	rounded := int(value + 0.5)

	return strconv.Itoa(rounded)
}
