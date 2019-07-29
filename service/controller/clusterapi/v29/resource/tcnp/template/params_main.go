package template

// ParamsMain is the data structure for the Tenant Cluster Node Pool template.
type ParamsMain struct {
	AutoScalingGroup    *ParamsMainAutoScalingGroup
	IAMPolicies         *ParamsMainIAMPolicies
	LaunchConfiguration *ParamsMainLaunchConfiguration
	LifecycleHooks      *ParamsMainLifecycleHooks
	Outputs             *ParamsMainOutputs
	SecurityGroups      *ParamsMainSecurityGroups
	Subnets             *ParamsMainSubnets
	VPCCIDR             *ParamsMainVPCCIDR
}
