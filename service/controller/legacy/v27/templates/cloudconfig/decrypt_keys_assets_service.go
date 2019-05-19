package cloudconfig

const DecryptKeysAssetsService = `
[Unit]
Description=Decrypt Keys assets
Before=k8s-kubelet.service
After=wait-for-domains.service vault-aws-authorizer.service
Requires=wait-for-domains.service vault-aws-authorizer.service

[Service]
Type=oneshot
ExecStart=/opt/bin/decrypt-keys-assets

[Install]
WantedBy=multi-user.target
`
