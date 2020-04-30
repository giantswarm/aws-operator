package template

type ParamsMainRecordSets struct {
	BaseDomain     string
	ClusterID      string
	HostedZoneID   string
	Records        []ParamsMainRecordSetsRecords
	Route53Enabled bool
}

type ParamsMainRecordSetsRecords struct {
	ResourceName    string
	ENIResourceName string
	Value           string
}
