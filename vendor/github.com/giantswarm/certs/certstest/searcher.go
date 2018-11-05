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

func (s *Searcher) SearchDraining(clusterID string) (certs.Draining, error) {
	return certs.Draining{}, nil
}

func (s *Searcher) SearchClusterOperator(clusterID string) (certs.ClusterOperator, error) {
	return certs.ClusterOperator{}, nil
}

func (s *Searcher) SearchMonitoring(clusterID string) (certs.Monitoring, error) {
	return certs.Monitoring{}, nil
}

func (s *Searcher) SearchTLS(clusterID string, cert certs.Cert) (certs.TLS, error) {
	return certs.TLS{}, nil
}
