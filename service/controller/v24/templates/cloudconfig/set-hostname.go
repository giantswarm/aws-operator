package cloudconfig

const SetHostname = `
[Unit]
Description=set proper hostname for k8s
Requires=wait-for-domains.service
After=wait-for-domains.service
Before=k8s-kubelet.service

[Service]
Type=oneshot
RemainAfterExit=yes
ExecStart=/bin/bash -c "hostnamectl set-hostname $(curl http://169.254.169.254/latest/meta-data/local-hostname)"

[Install]
WantedBy=multi-user.target
`
