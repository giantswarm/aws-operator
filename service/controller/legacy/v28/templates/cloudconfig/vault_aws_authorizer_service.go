package cloudconfig

const VaultAWSAuthorizerService = `
[Unit]
Description=Token decryption retrieval
Before=k8s-kubelet.service
After=wait-for-domains.service
Requires=wait-for-domains.service

[Service]
Type=oneshot
ExecStart=/opt/bin/vault-aws-authorizer
[Install]
WantedBy=multi-user.target
`
