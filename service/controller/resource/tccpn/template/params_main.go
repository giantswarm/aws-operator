package template

// ParamsMain is the data structure for the Tenant Cluster Control Plane Nodes
// template.
type ParamsMain struct {
	AutoScalingGroup    *ParamsMainAutoScalingGroup
	IAMPolicies         *ParamsMainIAMPolicies
	LaunchConfiguration *ParamsMainLaunchConfiguration
	Outputs             *ParamsMainOutputs
}
