package cloudconfig

const EphemeralVarLogMount = `
[Unit]
Description=log data volume
DefaultDependencies=no

[Mount]
What=/dev/disk/by-label/log
Where=/var/log
Type=ext4

[Install]
WantedBy=local-fs-pre.target
`
