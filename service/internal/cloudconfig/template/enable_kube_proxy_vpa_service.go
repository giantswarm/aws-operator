package template

const KubeProxyVPAService = `
[Unit]
Description=Enable VPA for kube-proxy
After=k8s-addons.service

[Service]
Type=oneshot
ExecStartPre=/bin/bash -c "while [ ! -f /etc/kubernetes/kubeconfig/addons.yaml ]; do echo 'Waiting for /etc/kubernetes/kubeconfig/addons.yaml to be written' && sleep 1; done"
ExecStartPre=/bin/bash -c "while ! /opt/bin/kubectl --kubeconfig=/etc/kubernetes/kubeconfig/addons.yaml get crd verticalpodautoscalers.autoscaling.k8s.io; do echo 'Waiting for VPA CRD to exists' && sleep 1; done"
ExecStart=/opt/bin/kubectl --kubeconfig=/etc/kubernetes/kubeconfig/addons.yaml apply -f /srv/kube-proxy-vpa.yaml

[Install]
WantedBy=multi-user.target
`
