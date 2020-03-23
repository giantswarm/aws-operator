package template

type ParamsMainAutoScalingGroup struct {
	AvailabilityZones     []string
	Cluster               ParamsMainAutoScalingGroupCluster
	DesiredCapacity       int
	LoadBalancers         ParamsMainAutoScalingGroupLoadBalancers
	MaxBatchSize          string
	MaxSize               int
	MinInstancesInService string
	MinSize               int
	Subnets               []string
}

type ParamsMainAutoScalingGroupCluster struct {
	ID string
}

type ParamsMainAutoScalingGroupLoadBalancers struct {
	ApiInternalName string
	ApiName         string
	EtcdName        string
}
