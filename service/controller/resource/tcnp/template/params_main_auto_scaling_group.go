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
	// OnDemandPercentageAboveBaseCapacity controls the percentages of On-Demand
	// Instances and Spot Instances for your additional capacity beyond
	// OnDemandBaseCapacity.
	OnDemandPercentageAboveBaseCapacity int
	// OnDemandBaseCapacity defines the minimum amount of the Auto Scaling group's
	// capacity that must be fulfilled by On-Demand Instances. This base portion is
	// provisioned first as your group scales.
	OnDemandBaseCapacity int
	// SpotAllocationStrategy If the allocation strategy is lowest-price, the Auto
	// Scaling group launches instances using the Spot pools with the lowest price,
	// and evenly allocates your instances across the number of Spot pools that you
	// specify. If the allocation strategy is capacity-optimized, the Auto Scaling
	// group launches instances using Spot pools that are optimally chosen based on
	// the available Spot capacity.
	SpotAllocationStrategy string
	// SpotInstancePools The number of Spot pools to use to allocate your Spot
	// capacity. The Spot pools are determined from the different instance types
	// in the Overrides array of LaunchTemplate. The range is 1–20. The default
	// value is 2.
	SpotInstancePools int
	// LaunchTemplateOverrides is an optional setting. Any parameters that you
	// specify override the same parameters in the launch template. Currently,
	// the only supported override is instance type. You can specify between 1 and
	// 20 instance types.
	LaunchTemplateOverrides []LaunchTemplateOverride
}

type ParamsMainAutoScalingGroupCluster struct {
	ID string
}

type LaunchTemplateOverride struct {
	InstanceType string
	// WeightedCapacity defines the number of capacity units, which gives the
	// instance type a proportional weight to other instance types. For example,
	// larger instance types are generally weighted more than smaller instance
	// types. These are the same units that you chose to set the desired capacity
	// in terms of instances, or a performance attribute such as vCPUs, memory, or I/O.
	WeightedCapacity int
}
