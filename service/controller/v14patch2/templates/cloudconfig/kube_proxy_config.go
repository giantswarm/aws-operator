package cloudconfig

const KubeProxyConfig = `apiVersion: kubeproxy.config.k8s.io/v1alpha1
clientConnection:
  kubeconfig: /etc/kubernetes/config/proxy-kubeconfig.yml
kind: KubeProxyConfiguration
mode: iptables
resourceContainer: /kube-proxy
`
