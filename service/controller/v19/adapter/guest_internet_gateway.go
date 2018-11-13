package adapter

import "github.com/giantswarm/aws-operator/service/controller/v19/key"

type GuestInternetGatewayAdapter struct {
	ClusterID string
}

func (a *GuestInternetGatewayAdapter) Adapt(cfg Config) error {
	a.ClusterID = key.ClusterID(cfg.CustomObject)

	return nil
}
