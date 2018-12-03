package adapter

type GuestNATGatewayAdapter struct {
	ClusterID string
}

func (a *GuestNATGatewayAdapter) Adapt(cfg Config) error {
	a.ClusterID = clusterID(cfg)

	return nil
}
