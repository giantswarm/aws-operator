package template

type ParamsMainAutoScalingGroup struct {
	List []ParamsMainAutoScalingGroupItem
}

type ParamsMainAutoScalingGroupItem struct {
	AvailabilityZone string
	ClusterID        string
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
