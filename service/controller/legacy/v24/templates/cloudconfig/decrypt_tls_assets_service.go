package cloudconfig

const DecryptTLSAssetsService = `
[Unit]
Description=Decrypt TLS certificates
Before=k8s-kubelet.service

[Service]
Type=oneshot
ExecStart=/opt/bin/decrypt-tls-assets

[Install]
WantedBy=multi-user.target
`
