package adapter

type GuestInternetGatewayAdapter struct {
	ClusterID string
}

func (a *GuestInternetGatewayAdapter) Adapt(cfg Config) error {
	a.ClusterID = clusterID(cfg)

	return nil
}
