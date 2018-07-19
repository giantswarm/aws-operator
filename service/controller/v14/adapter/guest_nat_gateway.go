package adapter

type guestNATGatewayAdapter struct {
	ClusterID string
}

func (a *guestNATGatewayAdapter) Adapt(cfg Config) error {
	a.ClusterID = clusterID(cfg)

	return nil
}
