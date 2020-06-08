package template

type ParamsMainAutoScalingGroup struct {
	List      []ParamsMainAutoScalingGroupItem
	HAMasters bool
}

type ParamsMainAutoScalingGroupItem struct {
	AvailabilityZone string
	ClusterID        string
	DependsOn        []string
	LaunchTemplate   ParamsMainAutoScalingGroupItemLaunchTemplate
	LoadBalancers    ParamsMainAutoScalingGroupItemLoadBalancers
	Resource         string
	SubnetID         string
}

type ParamsMainAutoScalingGroupItemLaunchTemplate struct {
	Resource string
}

type ParamsMainAutoScalingGroupItemLoadBalancers struct {
	ApiInternalName string
	ApiName         string
	EtcdName        string
}
