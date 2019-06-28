package template

type ParamsMainAutoScalingGroup struct {
	AvailabilityZones     []string
	Cluster               ParamsMainAutoScalingGroupCluster
	DesiredCapacity       int
	MaxBatchSize          string
	MaxSize               int
	MinInstancesInService string
	MinSize               int
	Name                  string
	Subnets               []string
}

type ParamsMainAutoScalingGroupCluster struct {
	ID string
}
