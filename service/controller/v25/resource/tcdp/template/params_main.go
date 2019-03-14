package template

// ParamsMain is the data structure for the Tenant Cluster Data Plane template.
type ParamsMain struct {
	AutoScalingGroup    *ParamsMainAutoScalingGroup
	IAMPolicies         *ParamsMainIAMPolicies
	LaunchConfiguration *ParamsMainLaunchConfig
	LifecycleHooks      *ParamsMainLifecycleHooks
	Outputs             *ParamsMainOutputs
	SecurityGroups      *ParamsMainSecurityGroups
	Subnets             *ParamsMainSubnets
}
