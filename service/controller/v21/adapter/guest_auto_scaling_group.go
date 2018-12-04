package adapter

import (
	"strconv"

	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/service/controller/v21/key"
)

type GuestAutoScalingGroupAdapter struct {
	ASGMaxSize             int
	ASGMinSize             int
	ASGType                string
	ClusterID              string
	HealthCheckGracePeriod int
	MaxBatchSize           string
	MinInstancesInService  string
	PrivateSubnets         []string
	RollingUpdatePauseTime string
	WorkerAZs              []string
}

func (a *GuestAutoScalingGroupAdapter) Adapt(cfg Config) error {
	workers := key.WorkerCount(cfg.CustomObject)
	if workers <= 0 {
		return microerror.Maskf(invalidConfigError, "at least 1 worker required, found %d", workers)
	}

	{
		numAZs := len(key.StatusAvailabilityZones(cfg.CustomObject))
		if numAZs < 1 {
			return microerror.Maskf(invalidConfigError, "at least one configured availability zone required")
		}
	}

	a.ASGMaxSize = workers + 1
	a.ASGMinSize = workers
	a.ASGType = key.KindWorker
	a.ClusterID = key.ClusterID(cfg.CustomObject)
	a.MaxBatchSize = workerCountRatio(workers, asgMaxBatchSizeRatio)
	a.MinInstancesInService = workerCountRatio(workers, asgMinInstancesRatio)
	a.HealthCheckGracePeriod = gracePeriodSeconds
	a.RollingUpdatePauseTime = rollingUpdatePauseTime

	for i, az := range key.StatusAvailabilityZones(cfg.CustomObject) {
		a.PrivateSubnets = append(a.PrivateSubnets, key.PrivateSubnetName(i))
		a.WorkerAZs = append(a.WorkerAZs, az.Name)
	}

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
