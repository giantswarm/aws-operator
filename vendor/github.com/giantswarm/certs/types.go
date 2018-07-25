package certs

type TLS struct {
	CA, Crt, Key []byte
}

type Cluster struct {
	APIServer        TLS
	CalicoClient     TLS
	CalicoEtcdClient TLS
	EtcdServer       TLS
	ServiceAccount   TLS
	Worker           TLS
}

type ClusterOperator struct {
	APIServer TLS
}

type Draining struct {
	NodeOperator TLS
}

type Monitoring struct {
	KubeStateMetrics TLS
	Prometheus       TLS
}
