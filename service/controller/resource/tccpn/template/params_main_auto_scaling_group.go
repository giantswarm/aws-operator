package template

type ParamsMainAutoScalingGroup struct {
	AvailabilityZone string
	ClusterID        string
	LoadBalancers    ParamsMainAutoScalingGroupLoadBalancers
	SubnetID         string
}

type ParamsMainAutoScalingGroupLoadBalancers struct {
	ApiInternalName string
	ApiName         string
	EtcdName        string
}
