package v1alpha2

import (
	"github.com/ghodss/yaml"
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	kindAWSCluster = "AWSCluster"
)

const awsClusterCRDYAML = `
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: awsclusters.infrastructure.giantswarm.io
spec:
  group: infrastructure.giantswarm.io
  names:
    kind: AWSCluster
    plural: awsclusters
    singular: awscluster
  scope: Namespaced
  subresources:
    status: {}
  validation:
    openAPIV3Schema:
      properties:
        spec:
          properties:
            cluster:
              properties:
                description:
                  maxLength: 100
                  type: string
                dns:
                  properties:
                    domain:
                      type: string
                    provider:
                      properties:
                        master:
                          properties:
                            availabilityZone:
                              type: string
                            instanceType:
                              type: string
                          type: object
                        region:
                          type: string
                      type: object
                  type: object
              type: object
          type: object
        status:
          properties:
            cluster:
              properties:
                conditions:
                  items:
                    properties:
                      lastTransitionTime:
                        format: date-time
                        type: string
                      type:
                        enum:
                          - Creating
                          - Created
                          - Updating
                          - Updated
                          - Deleting
                          - Deleted
                    type: object
                  type: array
                id:
                  pattern: "^[a-z0-9]{5}$"
                  type: string
                versions:
                  items:
                    properties:
                      lastTransitionTime:
                        format: date-time
                        type: string
                      version:
                        pattern: ^\d+\.\d+\.\d+$
                        type: string
                    type: object
                  type: array
              type: object
            provider:
              properties:
                network:
                  properties:
                    cidr:
                      pattern: ^\d+\.\d+\.\d+.\d+\/\d+$
                      type: string
                  type: object
              type: object
          type: object
  version: v1alpha2
`

var awsClusterCRD *apiextensionsv1beta1.CustomResourceDefinition

func init() {
	err := yaml.Unmarshal([]byte(awsClusterCRDYAML), &awsClusterCRD)
	if err != nil {
		panic(err)
	}
}

func NewAWSClusterCRD() *apiextensionsv1beta1.CustomResourceDefinition {
	return awsClusterCRD.DeepCopy()
}

func NewAWSClusterTypeMeta() metav1.TypeMeta {
	return metav1.TypeMeta{
		APIVersion: SchemeGroupVersion.String(),
		Kind:       kindAWSCluster,
	}
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// AWSCluster is the infrastructure provider referenced in upstream CAPI Cluster
// CRs.
//
//     apiVersion: infrastructure.giantswarm.io/v1alpha2
//     kind: AWSCluster
//     metadata:
//       labels:
//         aws-operator.giantswarm.io/version: 6.2.0
//         cluster-operator.giantswarm.io/version: 0.17.0
//         giantswarm.io/cluster: "8y5kc"
//         giantswarm.io/organization: "giantswarm"
//         release.giantswarm.io/version: 7.3.1
//       name: 8y5kc
//     spec:
//       cluster:
//         description: my fancy cluster
//         dns:
//           domain: gauss.eu-central-1.aws.gigantic.io
//         oidc:
//           claims:
//             username: email
//             groups: groups
//           clientID: foobar-dex-client
//           issuerURL: https://dex.gatekeeper.eu-central-1.aws.example.com
//       provider:
//         credentialSecret:
//           name: credential-default
//           namespace: giantswarm
//         master:
//           availabilityZone: eu-central-1a
//           instanceType: m4.large
//         region: eu-central-1
//     status:
//       cluster:
//         conditions:
//         - lastTransitionTime: "2019-03-25T17:10:09.333633991Z"
//           type: Created
//         id: 8y5kc
//         versions:
//         - lastTransitionTime: "2019-03-25T17:10:09.995948706Z"
//           version: 4.9.0
//       provider:
//         network:
//           cidr: 10.1.6.0/24
//
type AWSCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              AWSClusterSpec   `json:"spec" yaml:"spec"`
	Status            AWSClusterStatus `json:"status,omitempty" yaml:"status,omitempty"`
}

type AWSClusterSpec struct {
	Cluster  AWSClusterSpecCluster  `json:"cluster" yaml:"cluster"`
	Provider AWSClusterSpecProvider `json:"provider" yaml:"provider"`
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

type AWSClusterStatus struct {
	Cluster  CommonClusterStatus      `json:"cluster,omitempty" yaml:"cluster,omitempty"`
	Provider AWSClusterStatusProvider `json:"provider,omitempty" yaml:"provider,omitempty"`
}

type AWSClusterStatusProvider struct {
	Network AWSClusterStatusProviderNetwork `json:"network" yaml:"network"`
}

type AWSClusterStatusProviderNetwork struct {
	CIDR string `json:"cidr" yaml:"cidr"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type AWSClusterList struct {
	metav1.TypeMeta `json:",inline" yaml:",inline"`
	metav1.ListMeta `json:"metadata" yaml:"metadata"`
	Items           []AWSCluster `json:"items" yaml:"items"`
}
