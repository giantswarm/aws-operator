package cloudconfig

const EphemeralVarLogMount = `
[Unit]
Description=log data volume
Requires=format-var-log.service
After=format-var-log.service

[Mount]
What=/dev/disk/by-label/log
Where=/var/log
Type=ext4

[Install]
WantedBy=multi-user.target
`
