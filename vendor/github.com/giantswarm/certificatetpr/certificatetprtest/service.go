package certificatetprtest

import "github.com/giantswarm/certificatetpr"

type Service struct {
}

func NewService() Service {
	return Service{}
}

func (s Service) SearchCerts(clusterID string) (certificatetpr.AssetsBundle, error) {
	return certificatetpr.AssetsBundle{}, nil
}

func (s Service) SearchCertsForComponent(clusterID, componentName string) (certificatetpr.AssetsBundle, error) {
	return certificatetpr.AssetsBundle{}, nil
}
