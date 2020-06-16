apiVersion: infrastructure.giantswarm.io/v1alpha2
kind: AWSMachineDeployment
metadata:
  annotations:
    machine-deployment.giantswarm.io/subnet: 10.1.11.0/24
  creationTimestamp: "2020-06-11T09:17:40Z"
  finalizers:
  - operatorkit.giantswarm.io/aws-operator-machine-deployment-controller
  - operatorkit.giantswarm.io/aws-operator-drainer-controller
  generation: 1
  labels:
    aws-operator.giantswarm.io/version: 8.6.2-dev
    giantswarm.io/cluster: 8y5ck
    giantswarm.io/machine-deployment: ew9d7
    giantswarm.io/organization: giantswarm
    release.giantswarm.io/version: 11.5.0
  name: ew9d7
  namespace: default
  resourceVersion: "76395756"
  selfLink: /apis/infrastructure.giantswarm.io/v1alpha2/namespaces/default/awsmachinedeployments/ew9d7
  uid: de2eb28d-8640-490e-baaf-9071c0b8c6f5
spec:
  nodePool:
    description: 11.5.0 high availability
    machine:
      dockerVolumeSizeGB: 100
      kubeletVolumeSizeGB: 100
    scaling:
      max: 10
      min: 3
  provider:
    availabilityZones:
    - eu-central-1a
    instanceDistribution:
      onDemandBaseCapacity: 0
      onDemandPercentageAboveBaseCapacity: 100
    worker:
      instanceType: m4.xlarge
      useAlikeInstanceTypes: false
status:
  provider:
    worker:
      instanceTypes:
      - m4.xlarge
      spotInstances: 0
