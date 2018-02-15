package cloudconfig

const DecryptKeysAssetsServiceTemplate = `
[Unit]
Description=Decrypt Secret Keys

[Service]
Type=oneshot
ExecStart=/opt/bin/decrypt-keys-assets

[Install]
WantedBy=multi-user.target
`
