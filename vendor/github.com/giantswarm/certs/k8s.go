package certs

import "fmt"

// These constants are used when filtering the secrets, to only retrieve the
// ones we are interested in.
const (
	// certificateLabel is the label used in the secret to identify a secret
	// containing the certificate.
	certificateLabel = "giantswarm.io/certificate"
	// clusterLabel is the label used in the secret to identify a secret
	// containing the certificate.
	clusterLabel = "giantswarm.io/cluster"

	// legacyCertificateLabel is the label used in the secret to identify a secret
	// containing the certificate.
	//
	// TODO use certificateLabel instead when all cert secrets have it.
	legacyCertificateLabel = "clusterComponent"
	// legacyClusterIDLabel is the label used in the secret to identify a secret
	// containing the certificate.
	//
	// TODO use clusterIDLabel instead when all cert secrets have it.
	legacyClusterIDLabel = "clusterID"

	SecretNamespace = "default"
)

// Cert is a certificate name.
type Cert string

func (c Cert) String() string {
	return string(c)
}

// These constants used as Cert parsing a secret received from the API.
const (
	APICert                Cert = "api"
	AppOperatorAPICert     Cert = "app-operator-api"
	CalicoEtcdClientCert   Cert = "calico-etcd-client"
	ClusterOperatorAPICert Cert = "cluster-operator-api"
	EtcdCert               Cert = "etcd"
	FlanneldEtcdClientCert Cert = "flanneld-etcd-client"
	InternalAPICert        Cert = "internal-api"
	NodeOperatorCert       Cert = "node-operator"
	PrometheusCert         Cert = "prometheus"
	ServiceAccountCert     Cert = "service-account"
	WorkerCert             Cert = "worker"
)

// AllCerts lists all certificates that can be created by cert-operator.
var AllCerts = []Cert{
	APICert,
	AppOperatorAPICert,
	CalicoEtcdClientCert,
	ClusterOperatorAPICert,
	EtcdCert,
	FlanneldEtcdClientCert,
	InternalAPICert,
	NodeOperatorCert,
	PrometheusCert,
	ServiceAccountCert,
	WorkerCert,
}

// K8sName returns Kubernetes object name for the certificate name and
// the guest cluster ID.
func K8sName(clusterID string, certificate Cert) string {
	return fmt.Sprintf("%s-%s", clusterID, certificate)
}

// K8sLabels returns labels for the Kubernetes  object for the certificate name
// and the guest cluster ID.
func K8sLabels(clusterID string, certificate Cert) map[string]string {
	return map[string]string{
		certificateLabel:       string(certificate),
		clusterLabel:           clusterID,
		legacyCertificateLabel: string(certificate),
		legacyClusterIDLabel:   clusterID,
	}
}
