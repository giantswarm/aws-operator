package template

type ParamsMainAutoScalingGroup struct {
	AvailabilityZones                   []string
	Cluster                             ParamsMainAutoScalingGroupCluster
	DesiredCapacity                     int
	MaxBatchSize                        string
	MaxSize                             int
	MinInstancesInService               string
	MinSize                             int
	Subnets                             []string
	OnDemandPercentageAboveBaseCapacity int
	OnDemandBaseCapacity                int
	SpotAllocationStrategy              string
	LaunchTemplateOverrides             []LaunchTemplateOverride
}

type ParamsMainAutoScalingGroupCluster struct {
	ID string
}

type LaunchTemplateOverride struct {
	InstanceType     string
	WeightedCapacity string
}
