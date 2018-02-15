package cloudconfig

const IngressControllerConfigMapTemplate = `kind: ConfigMap
apiVersion: v1
metadata:
  name: ingress-nginx
  namespace: kube-system
  labels:
    k8s-addon: ingress-nginx.addons.k8s.io
data:
  server-name-hash-bucket-size: "1024"
  server-name-hash-max-size: "1024"
  use-proxy-protocol: "true"
`
