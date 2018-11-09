package adapter

import "github.com/giantswarm/aws-operator/service/controller/v19/key"

type GuestNATGatewayAdapter struct {
	ClusterID string
}

func (a *GuestNATGatewayAdapter) Adapt(cfg Config) error {
	a.ClusterID = key.ClusterID(cfg.CustomObject)

	return nil
}
