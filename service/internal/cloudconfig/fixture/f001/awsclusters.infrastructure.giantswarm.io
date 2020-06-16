apiVersion: infrastructure.giantswarm.io/v1alpha2
kind: AWSCluster
metadata:
  creationTimestamp: "2020-06-11T09:17:40Z"
  finalizers:
  - operatorkit.giantswarm.io/aws-operator-cluster-controller
  generation: 1
  labels:
    aws-operator.giantswarm.io/version: 8.6.2-dev
    giantswarm.io/cluster: 8y5ck
    giantswarm.io/organization: giantswarm
    release.giantswarm.io/version: 11.5.0
  name: 8y5ck
  namespace: default
  resourceVersion: "76393504"
  selfLink: /apis/infrastructure.giantswarm.io/v1alpha2/namespaces/default/awsclusters/8y5ck
  uid: b4823fdc-4003-4756-b484-167bfe3e7b29
spec:
  cluster:
    description: xh3b4sd
    dns:
      domain: gauss.eu-central-1.aws.gigantic.io
    oidc:
      claims: {}
  provider:
    credentialSecret:
      name: credential-default
      namespace: giantswarm
    master:
      availabilityZone: ""
      instanceType: ""
    pods:
      cidrBlock: 10.2.0.0/16
    region: eu-central-1
status:
  cluster:
    conditions:
    - condition: Creating
      lastTransitionTime: "2020-06-11T09:17:48Z"
    id: 8y5ck
  provider:
    network:
      cidr: 10.1.14.0/24
      vpcID: vpc-0cf56b90c717cf7c7
