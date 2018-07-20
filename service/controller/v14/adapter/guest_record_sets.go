package adapter

type GuestRecordSetsAdapter struct {
	BaseDomain                 string
	ClusterID                  string
	MasterInstanceResourceName string
	Route53Enabled             bool
}

func (a *GuestRecordSetsAdapter) Adapt(config Config) error {
	a.BaseDomain = baseDomain(config)
	a.ClusterID = clusterID(config)
	a.MasterInstanceResourceName = masterInstanceResourceName(config)
	a.Route53Enabled = route53Enabled(config)

	return nil
}
