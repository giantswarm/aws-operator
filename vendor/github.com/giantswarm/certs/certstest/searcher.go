package certstest

import (
	"github.com/giantswarm/certs"
)

type Config struct {
	AppOperator          certs.AppOperator
	AppOperatorError     error
	Cluster              certs.Cluster
	ClusterError         error
	ClusterOperator      certs.ClusterOperator
	ClusterOperatorError error
	Draining             certs.Draining
	DrainingError        error
	Monitoring           certs.Monitoring
	MonitoringError      error
	TLS                  certs.TLS
	TLSError             error
}

type Searcher struct {
	appOperator          certs.AppOperator
	appOperatorError     error
	cluster              certs.Cluster
	clusterError         error
	clusterOperator      certs.ClusterOperator
	clusterOperatorError error
	draining             certs.Draining
	drainingError        error
	monitoring           certs.Monitoring
	monitoringError      error
	tls                  certs.TLS
	tlsError             error
}

func NewSearcher(config Config) *Searcher {
	return &Searcher{
		appOperator:          config.AppOperator,
		appOperatorError:     config.AppOperatorError,
		cluster:              config.Cluster,
		clusterError:         config.ClusterError,
		clusterOperator:      config.ClusterOperator,
		clusterOperatorError: config.ClusterOperatorError,
		draining:             config.Draining,
		drainingError:        config.DrainingError,
		monitoring:           config.Monitoring,
		monitoringError:      config.MonitoringError,
		tls:                  config.TLS,
		tlsError:             config.TLSError,
	}
}

func (s *Searcher) SearchAppOperator(clusterID string) (certs.AppOperator, error) {
	if s.appOperatorError != nil {
		return certs.AppOperator{}, s.appOperatorError
	}

	return s.appOperator, nil
}

func (s *Searcher) SearchCluster(clusterID string) (certs.Cluster, error) {
	if s.clusterError != nil {
		return certs.Cluster{}, s.clusterError
	}

	return s.cluster, nil
}

func (s *Searcher) SearchClusterOperator(clusterID string) (certs.ClusterOperator, error) {
	if s.clusterOperatorError != nil {
		return certs.ClusterOperator{}, s.clusterOperatorError
	}

	return s.clusterOperator, nil
}

func (s *Searcher) SearchDraining(clusterID string) (certs.Draining, error) {
	if s.drainingError != nil {
		return certs.Draining{}, s.drainingError
	}

	return s.draining, nil
}

func (s *Searcher) SearchMonitoring(clusterID string) (certs.Monitoring, error) {
	if s.monitoringError != nil {
		return certs.Monitoring{}, s.monitoringError
	}

	return s.monitoring, nil
}

func (s *Searcher) SearchTLS(clusterID string, cert certs.Cert) (certs.TLS, error) {
	if s.tlsError != nil {
		return certs.TLS{}, s.tlsError
	}

	return s.tls, nil
}
