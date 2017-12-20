package certs

import "fmt"

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

	SecretNamespace = "default"
)

// Cert is a certificate name.
type Cert string

// These constants used as Cert parsing a secret received from the API.
const (
	APICert              Cert = "api"
	CalicoCert           Cert = "calico"
	EtcdCert             Cert = "etcd"
	FlanneldCert         Cert = "flanneld"
	KubeStateMetricsCert Cert = "kube-state-metrics"
	PrometheusCert       Cert = "prometheus"
	ServiceAccountCert   Cert = "service-account"
	WorkerCert           Cert = "worker"
)

// AllCerts lists all certificates that can be created by cert-operator.
var AllCerts = []Cert{
	APICert,
	CalicoCert,
	EtcdCert,
	FlanneldCert,
	KubeStateMetricsCert,
	PrometheusCert,
	ServiceAccountCert,
	WorkerCert,
}

// K8sSecretName returns Kubernetes Secret object name for the certificate name
// and the guest cluster ID.
func K8sSecretName(clusterID, certificate Cert) string {
	return fmt.Sprintf("%s-%s", clusterID, certificate)
}
