package template

type ParamsMainAutoScalingGroup struct {
	AvailabilityZones      []string
	Cluster                ParamsMainAutoScalingGroupCluster
	DesiredCapacity        int
	HealthCheckGracePeriod int
	MaxBatchSize           string
	MaxSize                int
	MinInstancesInService  string
	MinSize                int
	RollingUpdatePauseTime string
	Subnets                []string
}

type ParamsMainAutoScalingGroupCluster struct {
	ID string
}
