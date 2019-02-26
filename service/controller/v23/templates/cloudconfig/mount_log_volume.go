package cloudconfig

const EphemeralVarLogMount = `
[Unit]
Description=log data volume
DefaultDependencies=no

[Mount]
What=/dev/xvdf
Where=/var/log

[Install]
WantedBy=local-fs-pre.target
`
