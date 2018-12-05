package adapter

import (
	"strconv"

	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/service/controller/v21/key"
)

type GuestAutoScalingGroupAdapter struct {
	ASGDesiredCapacity     int
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
	maxWorkers := key.ScalingMax(cfg.CustomObject)
	minWorkers := key.ScalingMin(cfg.CustomObject)

	if minWorkers <= 0 {
		return microerror.Maskf(invalidConfigError, "at least 1 worker required, found %d", minWorkers)
	}

	if maxWorkers < minWorkers {
		return microerror.Maskf(invalidConfigError, "maximum number of workers (%d) is smaller than minimum number of workers (%d)", maxWorkers, minWorkers)
	}

	if maxWorkers == minWorkers {
		maxWorkers++
	}

	{
		numAZs := len(key.StatusAvailabilityZones(cfg.CustomObject))
		if numAZs < 1 {
			return microerror.Maskf(invalidConfigError, "at least one configured availability zone required")
		}
	}

	// Find out the minimum desired number of workers.
	currentDesiredMinWorkers := minDesiredWorkers(minWorkers, maxWorkers, key.StatusScalingDesiredCapacity(cfg.CustomObject))

	a.ASGDesiredCapacity = currentDesiredMinWorkers
	a.ASGMaxSize = maxWorkers
	a.ASGMinSize = minWorkers
	a.ASGType = key.KindWorker
	a.ClusterID = key.ClusterID(cfg.CustomObject)
	a.MaxBatchSize = workerCountRatio(currentDesiredMinWorkers, asgMaxBatchSizeRatio)
	a.MinInstancesInService = workerCountRatio(currentDesiredMinWorkers, asgMinInstancesRatio)
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

func minDesiredWorkers(minWorkers, maxWorkers, statusDesiredCapacity int) int {
	if statusDesiredCapacity > maxWorkers {
		return maxWorkers
	}

	if statusDesiredCapacity > minWorkers {
		return statusDesiredCapacity
	}

	return minWorkers
}
