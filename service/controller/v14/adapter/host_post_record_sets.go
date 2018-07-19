package adapter

type hostPostRecordSetsAdapter struct {
	BaseDomain                 string
	ClusterID                  string
	GuestHostedZoneNameServers string
	Route53Enabled             bool
}

func (r *hostPostRecordSetsAdapter) Adapt(config Config) error {
	r.BaseDomain = baseDomain(config)
	r.ClusterID = clusterID(config)
	r.Route53Enabled = route53Enabled(config)
	r.GuestHostedZoneNameServers = hostedZoneNameServers(config)

	return nil
}
