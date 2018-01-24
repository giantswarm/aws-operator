package certs

// These constants are used when filtering the secrets, to only retrieve the
// ones we are interested in.
const (
	// componentLabel is the label used in the secret to identify a secret
	// containing the certificate.
	//
	// TODO replace with "giantswarm.io/certificate" and add to
	// https://github.com/giantswarm/fmt.
	certficateLabel = "clusterComponent"
	// clusterIDLabel is the label used in the secret to identify a secret
	// containing the certificate.
	//
	// TODO replace with "giantswarm.io/cluster-id"
	clusterIDLabel = "clusterID"

	SecretNamesapce = "default"
)

type cert string

// These constants used as Cert
// parsing a secret received from the API.
const (
	apiCert              cert = "api"
	calicoCert           cert = "calico"
	etcdCert             cert = "etcd"
	flanneldCert         cert = "flanneld"
	kubeStateMetricsCert cert = "kube-state-metrics"
	prometheusCert       cert = "prometheus"
	serviceAccountCert   cert = "service-account"
	workerCert           cert = "worker"
)

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

type Monitoring struct {
	Prometheus       TLS
	KubeStateMetrics TLS
}
