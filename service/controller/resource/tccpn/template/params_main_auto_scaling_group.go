package template

type ParamsMainAutoScalingGroup struct {
	List      []ParamsMainAutoScalingGroupItem
	HAMasters bool
}

type ParamsMainAutoScalingGroupItem struct {
	AvailabilityZone string
	ClusterID        string
	Eni              ParamsMainAutoScalingGroupItemEni
	EtcdVolume       ParamsMainAutoScalingGroupItemEtcdVolume
	LaunchTemplate   ParamsMainAutoScalingGroupItemLaunchTemplate
	LoadBalancers    ParamsMainAutoScalingGroupItemLoadBalancers
	Resource         string
	SubnetID         string
}

type ParamsMainAutoScalingGroupItemEni struct {
	Resource string
}

type ParamsMainAutoScalingGroupItemEtcdVolume struct {
	Resource string
}

type ParamsMainAutoScalingGroupItemLaunchTemplate struct {
	Resource string
}

type ParamsMainAutoScalingGroupItemLoadBalancers struct {
	ApiInternalName string
	ApiName         string
	EtcdName        string
}
