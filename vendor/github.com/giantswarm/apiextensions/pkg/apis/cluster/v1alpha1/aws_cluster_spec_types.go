package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// AWSClusterSpec is the structure put into the provider spec of the Cluster
// API's Cluster type. There it is tracked as serialized raw extension.
//
//     kind: AWSClusterSpec
//     apiVersion: cluster.giantswarm.io/v1alpha1
//     metadata:
//       labels:
//         aws-operator.giantswarm.io/version: 6.2.0
//         cluster-operator.giantswarm.io/version: 0.17.0
//         giantswarm.io/cluster: "8y5kc"
//         giantswarm.io/organization: "giantswarm"
//         release.giantswarm.io/version: 7.3.1
//       name: 8y5kc
//     cluster:
//       description: my fancy cluster
//       dns:
//         domain: gauss.eu-central-1.aws.gigantic.io
//       oidc:
//         claims:
//           username: email
//           groups: groups
//         clientID: foobar-dex-client
//         issuerURL: https://dex.gatekeeper.eu-central-1.aws.example.com
//     provider:
//       credentialSecret:
//         name: credential-default
//         namespace: giantswarm
//       master:
//         availabilityZone: eu-central-1a
//         instanceType: m4.large
//       region: eu-central-1
//
type AWSClusterSpec struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Cluster           AWSClusterSpecCluster  `json:"cluster" yaml:"cluster"`
	Provider          AWSClusterSpecProvider `json:"provider" yaml:"provider"`
}

type AWSClusterSpecCluster struct {
	Description string                    `json:"description" yaml:"description"`
	DNS         AWSClusterSpecClusterDNS  `json:"dns" yaml:"dns"`
	OIDC        AWSClusterSpecClusterOIDC `json:"oidc" yaml:"oidc"`
}

type AWSClusterSpecClusterDNS struct {
	Domain string `json:"domain" yaml:"domain"`
}

type AWSClusterSpecClusterOIDC struct {
	Claims    AWSClusterSpecClusterOIDCClaims `json:"claims" yaml:"claims"`
	ClientID  string                          `json:"clientID" yaml:"clientID"`
	IssuerURL string                          `json:"issuerURL" yaml:"issuerURL"`
}

type AWSClusterSpecClusterOIDCClaims struct {
	Username string `json:"username" yaml:"username"`
	Groups   string `json:"groups" yaml:"groups"`
}

type AWSClusterSpecProvider struct {
	CredentialSecret AWSClusterSpecProviderCredentialSecret `json:"credentialSecret" yaml:"credentialSecret"`
	Master           AWSClusterSpecProviderMaster           `json:"master" yaml:"master"`
	Region           string                                 `json:"region" yaml:"region"`
}

type AWSClusterSpecProviderCredentialSecret struct {
	Name      string `json:"name" yaml:"name"`
	Namespace string `json:"namespace" yaml:"namespace"`
}

type AWSClusterSpecProviderMaster struct {
	AvailabilityZone string `json:"availabilityZone" yaml:"availabilityZone"`
	InstanceType     string `json:"instanceType" yaml:"instanceType"`
}
