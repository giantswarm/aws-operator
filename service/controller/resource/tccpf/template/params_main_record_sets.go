package template

type ParamsMainRecordSets struct {
	BaseDomain                     string
	ClusterID                      string
	ControlPlanePublicHostedZoneID string
	GuestHostedZoneNameServers     string
	Route53Enabled                 bool
}
