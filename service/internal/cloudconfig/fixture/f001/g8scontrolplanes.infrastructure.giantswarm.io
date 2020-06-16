apiVersion: infrastructure.giantswarm.io/v1alpha2
kind: G8sControlPlane
metadata:
  creationTimestamp: "2020-06-11T09:17:40Z"
  finalizers:
  - operatorkit.giantswarm.io/cluster-operator-control-plane-controller
  generation: 2
  labels:
    cluster-operator.giantswarm.io/version: 2.2.1-dev
    giantswarm.io/cluster: 8y5ck
    giantswarm.io/control-plane: 2dnhx
    giantswarm.io/organization: giantswarm
    release.giantswarm.io/version: 11.5.0
  name: 2dnhx
  namespace: default
  resourceVersion: "76392470"
  selfLink: /apis/infrastructure.giantswarm.io/v1alpha2/namespaces/default/g8scontrolplanes/2dnhx
  uid: 277ee8f7-2242-421c-b89e-b8744fc51ecb
spec:
  infrastructureRef:
    apiVersion: infrastructure.giantswarm.io/v1alpha2
    kind: AWSControlPlane
    name: 2dnhx
    namespace: default
    resourceVersion: "76392468"
    uid: 4802a0a1-068b-4532-94e4-eec45d01d883
  replicas: 3
