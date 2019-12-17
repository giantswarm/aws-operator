package v1alpha2

import (
	"github.com/ghodss/yaml"
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	kindAWSMachineDeployment = "AWSMachineDeployment"
)

const awsMachineDeploymentCRDYAML = `
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: awsmachinedeployments.infrastructure.giantswarm.io
spec:
  group: infrastructure.giantswarm.io
  names:
    kind: AWSMachineDeployment
    plural: awsmachinedeployments
    singular: awsmachinedeployment
  scope: Namespaced
  subresources:
    status: {}
  validation:
    openAPIV3Schema:
      properties:
        spec:
          properties:
            nodePool:
              properties:
                description:
                  maxLength: 100
                  type: string
                machine:
                  properties:
                    dockerVolumeSizeGB:
                      format: int32
                      type: integer
                    kubeletVolumeSizeGB:
                      format: int32
                      type: integer
                  type: object
                scaling:
                  properties:
                    max:
                      format: int32
                      type: integer
                    min:
                      format: int32
                      type: integer
                  type: object
              type: object
            provider:
              properties:
                availabilityZones:
                  items:
                    type: string
                  type: array
                worker:
                  properties:
                    instanceType:
                      type: string
              type: object
          type: object
  version: v1alpha2
`

var awsMachineDeploymentCRD *apiextensionsv1beta1.CustomResourceDefinition

func init() {
	err := yaml.Unmarshal([]byte(awsMachineDeploymentCRDYAML), &awsMachineDeploymentCRD)
	if err != nil {
		panic(err)
	}
}

func NewAWSMachineDeploymentCRD() *apiextensionsv1beta1.CustomResourceDefinition {
	return awsMachineDeploymentCRD.DeepCopy()
}

func NewAWSMachineDeploymentTypeMeta() metav1.TypeMeta {
	return metav1.TypeMeta{
		APIVersion: SchemeGroupVersion.String(),
		Kind:       kindAWSMachineDeployment,
	}
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// AWSMachineDeployment is the infrastructure provider referenced in upstream
// CAPI MachineDeployment CRs.
//
//     apiVersion: infrastructure.giantswarm.io/v1alpha2
//     kind: AWSMachineDeployment
//     metadata:
//       labels:
//         aws-operator.giantswarm.io/version: 6.2.0
//         cluster-operator.giantswarm.io/version: 0.17.0
//         giantswarm.io/cluster: 8y5kc
//         giantswarm.io/organization: "giantswarm"
//         giantswarm.io/machine-deployment: al9qy
//         release.giantswarm.io/version: 7.3.1
//       name: al9qy
//     spec:
//       nodePool:
//         description: my fancy node pool
//         machine:
//           dockerVolumeSizeGB: 100
//           kubeletVolumeSizeGB: 100
//         scaling:
//           max: 3
//           min: 3
//       provider:
//         availabilityZones:
//           - eu-central-1a
//         worker:
//           instanceType: m4.xlarge
//
type AWSMachineDeployment struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              AWSMachineDeploymentSpec `json:"spec" yaml:"spec"`
}

type AWSMachineDeploymentSpec struct {
	NodePool AWSMachineDeploymentSpecNodePool `json:"nodePool" yaml:"nodePool"`
	Provider AWSMachineDeploymentSpecProvider `json:"provider" yaml:"provider"`
}

type AWSMachineDeploymentSpecNodePool struct {
	Description string                                  `json:"description" yaml:"description"`
	Machine     AWSMachineDeploymentSpecNodePoolMachine `json:"machine" yaml:"machine"`
	Scaling     AWSMachineDeploymentSpecNodePoolScaling `json:"scaling" yaml:"scaling"`
}

type AWSMachineDeploymentSpecNodePoolMachine struct {
	DockerVolumeSizeGB  int `json:"dockerVolumeSizeGB" yaml:"dockerVolumeSizeGB"`
	KubeletVolumeSizeGB int `json:"kubeletVolumeSizeGB" yaml:"kubeletVolumeSizeGB"`
}

type AWSMachineDeploymentSpecNodePoolScaling struct {
	Max int `json:"max" yaml:"max"`
	Min int `json:"min" yaml:"min"`
}

type AWSMachineDeploymentSpecProvider struct {
	AvailabilityZones []string                               `json:"availabilityZones" yaml:"availabilityZones"`
	Worker            AWSMachineDeploymentSpecProviderWorker `json:"worker" yaml:"worker"`
}

type AWSMachineDeploymentSpecProviderWorker struct {
	InstanceType string `json:"instanceType" yaml:"instanceType"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type AWSMachineDeploymentList struct {
	metav1.TypeMeta `json:",inline" yaml:",inline"`
	metav1.ListMeta `json:"metadata" yaml:"metadata"`
	Items           []AWSMachineDeployment `json:"items" yaml:"items"`
}
