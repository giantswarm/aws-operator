package randomkeytprtest

import "github.com/giantswarm/randomkeytpr"

type Service struct {
}

func NewService() Service {
	return Service{}
}

func (s Service) SearchKeys(clusterID string) (map[randomkeytpr.Key][]byte, error) {
	return map[randomkeytpr.Key][]byte{}, nil
}

func (s Service) SearchKeysForKeytype(clusterID, keyType string) (map[randomkeytpr.Key][]byte, error) {
	return map[randomkeytpr.Key][]byte{}, nil
}
