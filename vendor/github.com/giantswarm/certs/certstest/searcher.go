package certstest

import (
	"github.com/giantswarm/certs"
)

type Searcher struct {
}

func NewSearcher() *Searcher {
	return &Searcher{}
}

func (s *Searcher) SearchCluster(clusterID string) (certs.Cluster, error) {
	return certs.Cluster{}, nil
}

func (s *Searcher) SearchMonitoring(clusterID string) (certs.Monitoring, error) {
	return certs.Monitoring{}, nil
}
