package template

const KubeProxyVPAYAML = `
apiVersion: autoscaling.k8s.io/v1
kind: VerticalPodAutoscaler
metadata:
  labels:
    app.kubernetes.io/name: kube-proxy
    giantswarm.io/service-type: system
    k8s-app: kube-proxy
    kubernetes.io/cluster-service: "true"
  name: kube-proxy
  namespace: kube-system
spec:
  resourcePolicy:
    containerPolicies:
    - containerName: kube-proxy
      controlledValues: RequestsOnly
      mode: Auto
  targetRef:
    apiVersion: apps/v1
    kind: DaemonSet
    name: kube-proxy
  updatePolicy:
    updateMode: Auto
`
