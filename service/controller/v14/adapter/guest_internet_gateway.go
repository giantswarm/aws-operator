package adapter

type guestInternetGatewayAdapter struct {
	ClusterID string
}

func (a *guestInternetGatewayAdapter) Adapt(cfg Config) error {
	a.ClusterID = clusterID(cfg)

	return nil
}
