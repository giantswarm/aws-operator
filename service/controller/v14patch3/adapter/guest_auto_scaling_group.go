package adapter

import (
	"strconv"

	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/service/controller/v14patch3/key"
)

type GuestAutoScalingGroupAdapter struct {
	ASGMaxSize             int
	ASGMinSize             int
	ASGType                string
	ClusterID              string
	HealthCheckGracePeriod int
	MaxBatchSize           string
	MinInstancesInService  string
	RollingUpdatePauseTime string
	WorkerAZ               string
}

func (a *GuestAutoScalingGroupAdapter) Adapt(cfg Config) error {
	workers := key.WorkerCount(cfg.CustomObject)
	if workers <= 0 {
		return microerror.Maskf(invalidConfigError, "at least 1 worker required, found %d", workers)
	}

	a.WorkerAZ = key.AvailabilityZone(cfg.CustomObject)
	a.ASGMaxSize = workers + 1
	a.ASGMinSize = workers
	a.ASGType = asgType(cfg)
	a.ClusterID = clusterID(cfg)
	a.MaxBatchSize = workerCountRatio(workers, asgMaxBatchSizeRatio)
	a.MinInstancesInService = workerCountRatio(workers, asgMinInstancesRatio)
	a.HealthCheckGracePeriod = gracePeriodSeconds
	a.RollingUpdatePauseTime = rollingUpdatePauseTime

	return nil
}

func workerCountRatio(workers int, ratio float32) string {
	value := float32(workers) * ratio
	rounded := int(value + 0.5)

	if rounded == 0 {
		rounded = 1
	}

	return strconv.Itoa(rounded)
}
