package template

// ParamsMain is the data structure for the Tenant Cluster Control Plane Nodes
// template.
type ParamsMain struct {
	AutoScalingGroup    *ParamsMainAutoScalingGroup
	EtcdVolume          *ParamsMainEtcdVolume
	IAMPolicies         *ParamsMainIAMPolicies
	LaunchConfiguration *ParamsMainLaunchConfiguration
	Outputs             *ParamsMainOutputs
	RouteTables         *ParamsMainRouteTables
	SecurityGroups      *ParamsMainSecurityGroups
	Subnets             *ParamsMainSubnets
}
