package template

// ParamsMain is the data structure for the Tenant Cluster Control Plane Nodes
// template.
type ParamsMain struct {
	AutoScalingGroup *ParamsMainAutoScalingGroup
	ENI              *ParamsMainENI
	EtcdVolume       *ParamsMainEtcdVolume
	IAMPolicies      *ParamsMainIAMPolicies
	LaunchTemplate   *ParamsMainLaunchTemplate
	RecordSets       *ParamsMainRecordSets
	Outputs          *ParamsMainOutputs
}
