package template

type ParamsMainAutoScalingGroup struct {
	ASG                    ParamsMainAutoScalingGroupASG
	AvailabilityZones      []string
	Cluster                ParamsMainAutoScalingGroupCluster
	HealthCheckGracePeriod int
	MaxBatchSize           string
	MinInstancesInService  string
	RollingUpdatePauseTime string
	Subnets                []string
}

type ParamsMainAutoScalingGroupASG struct {
	DesiredCapacity int
	MaxSize         int
	MinSize         int
	Type            string
}

type ParamsMainAutoScalingGroupCluster struct {
	ID string
}
