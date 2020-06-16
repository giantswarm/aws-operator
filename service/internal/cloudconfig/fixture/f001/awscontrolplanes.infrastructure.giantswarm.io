apiVersion: infrastructure.giantswarm.io/v1alpha2
kind: AWSControlPlane
metadata:
  creationTimestamp: "2020-06-11T09:17:40Z"
  finalizers:
  - operatorkit.giantswarm.io/aws-operator-control-plane-controller
  - operatorkit.giantswarm.io/aws-operator-drainer-controller
  generation: 1
  labels:
    aws-operator.giantswarm.io/version: 8.6.2-dev
    giantswarm.io/cluster: 8y5ck
    giantswarm.io/control-plane: 2dnhx
    giantswarm.io/organization: giantswarm
    release.giantswarm.io/version: 11.5.0
  name: 2dnhx
  namespace: default
  resourceVersion: "76392480"
  selfLink: /apis/infrastructure.giantswarm.io/v1alpha2/namespaces/default/awscontrolplanes/2dnhx
  uid: 4802a0a1-068b-4532-94e4-eec45d01d883
spec:
  availabilityZones:
  - eu-central-1b
  - eu-central-1c
  - eu-central-1a
  instanceType: m5.xlarge
