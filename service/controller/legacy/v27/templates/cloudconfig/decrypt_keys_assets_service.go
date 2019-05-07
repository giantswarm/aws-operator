package cloudconfig

const DecryptKeysAssetsService = `
[Unit]
Description=Decrypt Secret Keys
Before=k8s-kubelet.service
After=wait-for-domains.service
Requires=wait-for-domains.service

[Service]
Type=oneshot
Restart=on-failure
RestartSec=30s
ExecStart=/opt/bin/decrypt-keys-assets

[Install]
WantedBy=multi-user.target
`
