package adapter

import (
	"strconv"

	"github.com/giantswarm/microerror"

	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v27/key"
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
	maxWorkers := key.WorkerScalingMax(cfg.MachineDeployment)
	minWorkers := key.WorkerScalingMin(cfg.MachineDeployment)

	if minWorkers <= 0 {
		return microerror.Maskf(invalidConfigError, "at least 1 worker required, found %d", minWorkers)
	}

	if maxWorkers < minWorkers {
		return microerror.Maskf(invalidConfigError, "maximum number of workers (%d) is smaller than minimum number of workers (%d)", maxWorkers, minWorkers)
	}

	{
		numAZs := len(key.StatusAvailabilityZones(cfg.MachineDeployment))
		if numAZs < 1 {
			return microerror.Maskf(invalidConfigError, "at least one configured availability zone required")
		}
	}

	// Find out the minimum desired number of workers.
	currentDesiredMinWorkers := minDesiredWorkers(minWorkers, maxWorkers, cfg.StackState.WorkerDesired)

	a.ASGDesiredCapacity = currentDesiredMinWorkers
	a.ASGMaxSize = maxWorkers
	a.ASGMinSize = minWorkers
	a.ASGType = "worker"
	a.ClusterID = key.ClusterID(cfg.CustomObject)
	a.MaxBatchSize = strconv.Itoa(workerCountRatio(currentDesiredMinWorkers, asgMaxBatchSizeRatio))

	minInstancesInService := workerCountRatio(currentDesiredMinWorkers, asgMinInstancesRatio)
	if minWorkers == 1 && maxWorkers == 1 {
		// MinInstancesInService must be less than the autoscaling group's MaxSize.
		// This should only ever be an issue if the min and max workers are exactly one.
		// If this is the case then the cluster will most likely go down while the new worker
		// is coming up and the old one is removed.
		minInstancesInService = 0
	}

	a.MinInstancesInService = strconv.Itoa(minInstancesInService)
	a.HealthCheckGracePeriod = gracePeriodSeconds
	a.RollingUpdatePauseTime = rollingUpdatePauseTime

	for i, az := range key.StatusAvailabilityZones(cfg.MachineDeployment) {
		a.PrivateSubnets = append(a.PrivateSubnets, key.PrivateSubnetName(i))
		a.WorkerAZs = append(a.WorkerAZs, az.Name)
	}

	return nil
}

func workerCountRatio(workers int, ratio float32) int {
	value := float32(workers) * ratio
	rounded := int(value + 0.5)

	if rounded == 0 {
		rounded = 1
	}

	return rounded
}

// minDesiredWorkers calculates appropriate minimum value to be set for ASG
// Desired value and to be used for computation of workerCountRatio.
//
// When cluster-autoscaler has scaled cluster and ASG's Desired value is higher
// than minimum number of instances allowed for that ASG, then it makes sense to
// consider Desired value as minimum number of running instances for further
// operational computations.
//
// Example:
// Initially ASG has minimum of 3 workers and maximum of 10. Due to amount of
// workload deployed on workers, cluster-autoscaler has scaled current Desired
// number of instances to 5. Therefore it makes sense to consider 5 as minimum
// number of nodes also when working on batch updates on ASG instances.
//
// Example 2:
// When end user is scaling cluster and adding restrictions to its size, it
// might be that initial ASG configuration is following:
// 		- Min: 3
//		- Max: 10
// 		- Desired: 10
//
// Now end user decides that it must be scaled down so maximum size is decreased
// to 7. When desired number of instances is temporarily bigger than maximum
// number of instances, it must be fixed to be maximum number of instances.
//
func minDesiredWorkers(minWorkers, maxWorkers, statusDesiredCapacity int) int {
	if statusDesiredCapacity > maxWorkers {
		return maxWorkers
	}

	if statusDesiredCapacity > minWorkers {
		return statusDesiredCapacity
	}

	return minWorkers
}
