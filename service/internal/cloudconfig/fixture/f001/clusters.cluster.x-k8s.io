apiVersion: cluster.x-k8s.io/v1alpha2
kind: Cluster
metadata:
  creationTimestamp: "2020-06-11T09:17:40Z"
  finalizers:
  - operatorkit.giantswarm.io/cluster-operator-cluster-controller
  - operatorkit.giantswarm.io/clusterapi-controller
  generation: 1
  labels:
    cluster-operator.giantswarm.io/version: 2.2.1-dev
    giantswarm.io/cluster: 8y5ck
    giantswarm.io/organization: giantswarm
    release.giantswarm.io/version: 11.5.0
  name: 8y5ck
  namespace: default
  resourceVersion: "76392647"
  selfLink: /apis/cluster.x-k8s.io/v1alpha2/namespaces/default/clusters/8y5ck
  uid: 1dcaf39b-8378-40f2-ab31-9be775e2cdfc
spec:
  infrastructureRef:
    apiVersion: infrastructure.giantswarm.io/v1alpha2
    kind: AWSCluster
    name: 8y5ck
    namespace: default
    resourceVersion: "76392464"
    uid: b4823fdc-4003-4756-b484-167bfe3e7b29
status:
  apiEndpoints:
  - host: api.8y5ck.k8s.gauss.eu-central-1.aws.gigantic.io
    port: 443
  controlPlaneInitialized: false
  infrastructureReady: false
