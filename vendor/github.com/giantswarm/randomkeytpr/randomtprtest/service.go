package randomkeytprtest

import "github.com/giantswarm/randomkeytpr"

func keySample() map[randomkeytpr.Key][]byte {
	r := make(map[randomkeytpr.Key][]byte)
	r[randomkeytpr.EncryptionKey] = []byte("+2cYxG60XQkggsdn4bX1neWrtJt60wJO2OXqmfOTBGc=")
	return r
}

type Service struct {
}

func NewService() Service {
	return Service{}
}

func (s Service) SearchKeys(clusterID string) (map[randomkeytpr.Key][]byte, error) {
	return keySample(), nil
}

func (s Service) SearchKeysForKeytype(clusterID, keyType string) (map[randomkeytpr.Key][]byte, error) {
	return keySample(), nil
}
