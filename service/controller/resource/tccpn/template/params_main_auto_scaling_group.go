package template

type ParamsMainAutoScalingGroup struct {
	AvailabilityZone string
	ClusterID        string
	LoadBalancers    ParamsMainAutoScalingGroupLoadBalancers
	ResourceNames    []string
	SubnetID         string
}

type ParamsMainAutoScalingGroupLoadBalancers struct {
	ApiInternalName string
	ApiName         string
	EtcdName        string
}
