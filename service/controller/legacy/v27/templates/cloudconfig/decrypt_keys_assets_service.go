package cloudconfig

const DecryptKeysAssetsService = `
[Unit]
Description=Decrypt Keys assets
Before=k8s-kubelet.service
After=decrypt-tls-assets.service
Requires=decrypt-tls-assets.service

[Service]
Type=oneshot
ExecStart=/opt/bin/decrypt-keys-assets

[Install]
WantedBy=multi-user.target
`
