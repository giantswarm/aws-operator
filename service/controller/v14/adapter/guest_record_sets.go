package adapter

type guestRecordSetsAdapter struct {
	BaseDomain                 string
	ClusterID                  string
	MasterInstanceResourceName string
	Route53Enabled             bool
}

func (a *guestRecordSetsAdapter) Adapt(config Config) error {
	a.BaseDomain = baseDomain(config)
	a.ClusterID = clusterID(config)
	a.MasterInstanceResourceName = masterInstanceResourceName(config)
	a.Route53Enabled = route53Enabled(config)

	return nil
}
