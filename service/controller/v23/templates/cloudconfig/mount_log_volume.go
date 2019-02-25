package cloudconfig

const EphemeralVarLogMount = `
[Unit]
Description=log data volume
Before=local-fs.target

[Mount]
What=/dev/disk/by-label/log
Where=/var/log
Type=ext4

[Install]
WantedBy=local-fs.target
`
