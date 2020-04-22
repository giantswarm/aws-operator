package template

// ParamsMain is the data structure for the Tenant Cluster Node Pool template.
type ParamsMain struct {
	AutoScalingGroup *ParamsMainAutoScalingGroup
	IAMPolicies      *ParamsMainIAMPolicies
	LaunchTemplate   *ParamsMainLaunchTemplate
	Outputs          *ParamsMainOutputs
	RouteTables      *ParamsMainRouteTables
	SecurityGroups   *ParamsMainSecurityGroups
	Subnets          *ParamsMainSubnets
	VPC              *ParamsMainVPC
}
