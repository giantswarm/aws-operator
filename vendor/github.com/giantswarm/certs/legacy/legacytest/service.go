package legacytest

import (
	"github.com/giantswarm/certs/legacy"
)

type Service struct {
}

func NewService() Service {
	return Service{}
}

func (s Service) SearchCerts(clusterID string) (legacy.AssetsBundle, error) {
	return legacy.AssetsBundle{}, nil
}

func (s Service) SearchCertsForComponent(clusterID, componentName string) (legacy.AssetsBundle, error) {
	return legacy.AssetsBundle{}, nil
}
