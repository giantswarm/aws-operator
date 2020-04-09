package template

type ParamsMainAutoScalingGroup struct {
	AvailabilityZones     []string
	Cluster               ParamsMainAutoScalingGroupCluster
	DesiredCapacity       int
	LifeCycleHookName     string
	MaxBatchSize          string
	MaxSize               int
	MinInstancesInService string
	MinSize               int
	Subnets               []string
}

type ParamsMainAutoScalingGroupCluster struct {
	ID string
}
