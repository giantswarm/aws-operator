package cloudconfig

const DecryptKeysAssetsService = `
[Unit]
Description=Decrypt Secret Keys
Before=k8s-kubelet.service

[Service]
Type=oneshot
ExecStart=/opt/bin/decrypt-keys-assets

[Install]
WantedBy=multi-user.target
`
