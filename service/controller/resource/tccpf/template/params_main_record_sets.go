package template

type ParamsMainRecordSets struct {
	BaseDomain                       string
	ClusterID                        string
	ControlPlaneInternalHostedZoneID string
	ControlPlaneHostedZoneID         string
	TenantAPIPublicLoadBalancer      string
	TenantHostedZoneNameServers      string
	Route53Enabled                   bool
}
