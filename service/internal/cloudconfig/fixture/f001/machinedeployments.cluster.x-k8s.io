apiVersion: cluster.x-k8s.io/v1alpha2
kind: MachineDeployment
metadata:
  creationTimestamp: "2020-06-11T09:17:40Z"
  finalizers:
  - operatorkit.giantswarm.io/cluster-operator-machine-deployment-controller
  generation: 1
  labels:
    cluster-operator.giantswarm.io/version: 2.2.1-dev
    giantswarm.io/cluster: 8y5ck
    giantswarm.io/machine-deployment: ew9d7
    giantswarm.io/organization: giantswarm
    release.giantswarm.io/version: 11.5.0
  name: ew9d7
  namespace: default
  resourceVersion: "76392478"
  selfLink: /apis/cluster.x-k8s.io/v1alpha2/namespaces/default/machinedeployments/ew9d7
  uid: cfc2b5ee-fdff-4317-9c3c-bbfd147f4eef
spec:
  selector: {}
  template:
    metadata: {}
    spec:
      bootstrap: {}
      infrastructureRef:
        apiVersion: infrastructure.giantswarm.io/v1alpha2
        kind: AWSMachineDeployment
        name: ew9d7
        namespace: default
        resourceVersion: "76392474"
        uid: de2eb28d-8640-490e-baaf-9071c0b8c6f5
      metadata: {}
