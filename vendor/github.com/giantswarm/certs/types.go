package certs

type TLS struct {
	CA, Crt, Key []byte
}

type Cluster struct {
	APIServer      TLS
	Worker         TLS
	ServiceAccount TLS
	CalicoClient   TLS
	EtcdServer     TLS
}

type Draining struct {
	NodeOperator TLS
}

type Monitoring struct {
	Prometheus       TLS
	KubeStateMetrics TLS
}
