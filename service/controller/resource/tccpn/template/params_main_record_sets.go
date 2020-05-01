package template

type ParamsMainRecordSets struct {
	BaseDomain     string
	ClusterID      string
	HostedZoneID   string
	Records        []ParamsMainRecordSetsRecord
	Route53Enabled bool
}

type ParamsMainRecordSetsRecord struct {
	ENI      ParamsMainRecordSetsRecordENI
	Resource string
	Value    string
}

type ParamsMainRecordSetsRecordENI struct {
	Resource string
}
